package main

import (
	"context"
	"log"
	"os"

	"sync"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
	"golang.org/x/sync/errgroup"
)

const (
	numWorkers = 10 // Number of concurrent workers to update subscriptions
)

// priceMapping maps the old price IDs to the new price IDs
var priceMapping = map[string]string{
	// Fill with your data
}

// This script will migrate all active subscriptions in Stripe from old prices to new prices using goroutines for performance.
// It first fetches all subscriptions that need to be migrated per price mapping,
// and then updates them, each operation performing its tasks in parallel.
// (i.e. all subscriptions are fetched in parallel, then all updates are performed in parallel).
//
// The script is thread-safe.
// Waitgroups are used for the fetching, and semaphores and channels are used for the updates.
func main() {
	stripe.Key = os.Getenv("STRIPE_API_KEY")
	if stripe.Key == "" {
		log.Fatal("STRIPE_API_KEY environment variable is required")
	}

	log.Printf("Starting migration for %d price mappings", len(priceMapping))

	subscriptionsToMigrate := make(map[string]SubscriptionUpdateParams) // The entire list of subscriptions to migrate across all price mappings
	priceStats := make(map[string]*PriceMigrationStats)

	var fetchWaitGroup sync.WaitGroup
	for oldPriceID, newPriceID := range priceMapping {
		priceStats[oldPriceID] = &PriceMigrationStats{
			OldPriceID:    oldPriceID,
			NewPriceID:    newPriceID,
			FailedSubsIDs: make([]string, 0),
		}
		fetchWaitGroup.Add(1)
		go fetchSubscriptionsToMigrate(oldPriceID, newPriceID, priceStats, subscriptionsToMigrate, &fetchWaitGroup)
	}

	fetchWaitGroup.Wait()

	if len(subscriptionsToMigrate) == 0 {
		log.Println("No subscriptions found to migrate")
		return
	}

	log.Printf("Total subscriptions to migrate: %d", len(subscriptionsToMigrate))

	// Channel to collect migration results
	results := make(chan MigrationResult, len(subscriptionsToMigrate))
	resultAggregationDoneChannel := make(chan struct{})

	// Collect results from the update goroutines by listening to the `results` channel.
	// This is a separate goroutine as well to avoid blocking the main thread
	go func() {
		defer close(resultAggregationDoneChannel)

		for i := 0; i < len(subscriptionsToMigrate); i++ {
			result := <-results

			stats := priceStats[result.OldPriceID]
			if !result.Success {
				stats.FailureCount++
				stats.FailedSubsIDs = append(stats.FailedSubsIDs, result.SubscriptionID)
				log.Printf("✗ Failed to migrate subscription %s from %s to %s: %v",
					result.SubscriptionID, result.OldPriceID, result.NewPriceID, result.Error)
				continue
			}

			stats.SuccessCount++
			log.Printf("✓ Migrated subscription %s from %s to %s",
				result.SubscriptionID, result.OldPriceID, result.NewPriceID)
		}
	}()

	// Update prices in parallel, leveraging errgroup to handle errors
	// and limiting the number of concurrent goroutines to numWorkers.
	errorGroup, ctx := errgroup.WithContext(context.Background())
	semaphore := make(chan struct{}, numWorkers)

	for _, migrationJob := range subscriptionsToMigrate {
		migrationJob := migrationJob // Necessary to capture loop variable
		errorGroup.Go(func() error {
			select {
			case semaphore <- struct{}{}:
				// Release the semaphore to allow another goroutine to run once finished
				defer func() {
					<-semaphore
				}()
			case <-ctx.Done():
				return ctx.Err()
			}

			_, err := subscription.Update(
				migrationJob.SubscriptionID,
				&stripe.SubscriptionParams{
					ProrationBehavior: stripe.String(string(stripe.SubscriptionSchedulePhaseProrationBehaviorNone)),
					Items: []*stripe.SubscriptionItemsParams{{
						ID:    stripe.String(migrationJob.ItemID),
						Price: stripe.String(migrationJob.NewPriceID),
					}},
				},
			)

			// Send result to aggregator channel
			results <- MigrationResult{
				SubscriptionID: migrationJob.SubscriptionID,
				OldPriceID:     migrationJob.OldPriceID,
				NewPriceID:     migrationJob.NewPriceID,
				Success:        err == nil,
				Error:          err,
			}

			return nil
		})
	}

	if err := errorGroup.Wait(); err != nil {
		log.Fatalf("bulk migration error: %v", err)
	}

	// Wait for result aggregation to complete
	<-resultAggregationDoneChannel

	// Report results per price pair
	log.Println("\n=== MIGRATION RESULTS ===")
	totalSuccess := 0
	totalFailure := 0
	totalProcessed := 0

	for oldPriceID, stats := range priceStats {
		if stats.TotalCount == 0 {
			continue
		}

		successRate := float64(stats.SuccessCount) / float64(stats.TotalCount) * 100
		log.Printf("%s -> %s:", oldPriceID, stats.NewPriceID)
		log.Printf("  Total: %d, Success: %d (%.1f%%), Failed: %d (%.1f%%)",
			stats.TotalCount, stats.SuccessCount, successRate,
			stats.FailureCount, float64(stats.FailureCount)/float64(stats.TotalCount)*100)

		if len(stats.FailedSubsIDs) > 0 {
			log.Printf("  Failed subscription IDs: %v", stats.FailedSubsIDs)
		}

		totalSuccess += stats.SuccessCount
		totalFailure += stats.FailureCount
		totalProcessed += stats.TotalCount
	}

	// Report overall results
	log.Println("\n=== OVERALL RESULTS ===")
	log.Printf("Total subscriptions processed: %d", totalProcessed)
	log.Printf("Successful migrations: %d (%.2f%%)", totalSuccess, float64(totalSuccess)/float64(totalProcessed)*100)
	log.Printf("Failed migrations: %d (%.2f%%)", totalFailure, float64(totalFailure)/float64(totalProcessed)*100)
	log.Printf("Price mappings processed: %d", len(priceMapping))

	if totalFailure > 0 {
		log.Printf("\nReview the failed subscriptions above for manual intervention.")
	} else {
		log.Printf("🎉 All subscriptions migrated successfully!")
	}
}

// fetchSubscriptionsToMigrate fetches all active subscriptions that use a specific price ID (`oldPriceID`)
// and adds them to the `subscriptionsToMigrate` map.
//
// It also updates the `priceStats` for the `oldPriceID`.
//
// It assumes that each subscription has only one item, or that the last one is the one we want to migrate.
func fetchSubscriptionsToMigrate(oldPriceID, newPriceID string, priceStats map[string]*PriceMigrationStats, subscriptionsToMigrate map[string]SubscriptionUpdateParams, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Fetching subscriptions for price %s -> %s", oldPriceID, newPriceID)

	iter := subscription.List(&stripe.SubscriptionListParams{
		Status: stripe.String("active"),
		Price:  stripe.String(oldPriceID),
	})

	count := 0
	for iter.Next() {
		s := iter.Subscription()
		// NOTE: We assume that each subscription has only one item, or that the last one is the one we want to migrate
		item := s.Items.Data[len(s.Items.Data)-1]

		subscriptionsToMigrate[s.ID] = SubscriptionUpdateParams{
			SubscriptionID: s.ID,
			ItemID:         item.ID,
			OldPriceID:     oldPriceID,
			NewPriceID:     newPriceID,
		}
		count++
	}

	if err := iter.Err(); err != nil {
		log.Fatalf("list error for price %s: %v", oldPriceID, err)
	}

	priceStats[oldPriceID].TotalCount = count
	log.Printf("Found %d subscriptions for price %s", count, oldPriceID)
}
