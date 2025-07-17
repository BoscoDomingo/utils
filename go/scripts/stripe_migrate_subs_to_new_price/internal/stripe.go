package internal

import (
	"fmt"
	"log"
	"sync"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/subscription"
)

// getPriceIDsFromLookupKeys fetches the price IDs from the lookup keys and appends them to priceIDMapping.
func getPriceIDsFromLookupKeys(lookupKeyMapping map[string]string, priceIDMapping map[string]string) {
	log.Printf("Resolving %d lookup key pairs to price IDs", len(lookupKeyMapping))

	type priceIDPair struct {
		oldPriceID string
		newPriceID string
		err        error
	}

	resultsChan := make(chan priceIDPair, len(lookupKeyMapping))
	var wg sync.WaitGroup

	for oldLookupKey, newLookupKey := range lookupKeyMapping {
		wg.Add(1)
		go func(oldKey, newKey string) {
			defer wg.Done()

			oldPriceID, err := getPriceIDByLookupKey(oldKey)
			if err != nil {
				resultsChan <- priceIDPair{err: fmt.Errorf("failed to resolve old lookup key %s: %w", oldKey, err)}
				return
			}

			newPriceID, err := getPriceIDByLookupKey(newKey)
			if err != nil {
				resultsChan <- priceIDPair{err: fmt.Errorf("failed to resolve new lookup key %s: %w", newKey, err)}
				return
			}

			resultsChan <- priceIDPair{
				oldPriceID: oldPriceID,
				newPriceID: newPriceID,
			}
		}(oldLookupKey, newLookupKey)
	}

	wg.Wait()
	close(resultsChan)

	for result := range resultsChan {
		if result.err != nil {
			log.Fatalf("Price lookup error: %v", result.err)
		}

		priceIDMapping[result.oldPriceID] = result.newPriceID
	}

	log.Printf("Successfully resolved %d lookup key pairs and added to price ID mapping", len(lookupKeyMapping))
}

// getPriceIDByLookupKey resolves a single lookup key to its price ID
func getPriceIDByLookupKey(lookupKey string) (string, error) {
	iter := price.List(&stripe.PriceListParams{
		LookupKeys: []*string{stripe.String(lookupKey)},
	})

	for iter.Next() {
		p := iter.Price()
		if p.LookupKey == lookupKey {
			return p.ID, nil
		}
	}

	if err := iter.Err(); err != nil {
		return "", err
	}

	// If we get here, no price was found for the lookup key
	return "", fmt.Errorf("no price found for lookup key: %s", lookupKey)
}

// getSubscriptionsToMigrate fetches all active subscriptions that use a specific price ID
// and adds them to the `subscriptionsToMigrate` map with the new price ID.
//
// It also updates the `priceStats` for the `oldPriceID`.
//
// It assumes that each subscription has only one item, or that the last one is the one we want to migrate.
func getSubscriptionsToMigrate(oldPriceID, newPriceID string, priceStats map[string]*PriceMigrationStats, subscriptionsToMigrate map[string]SubscriptionUpdateParams, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Fetching subscriptions for price ID %s", oldPriceID)

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
		log.Fatalf("list error for price ID %s: %v", oldPriceID, err)
	}

	priceStats[oldPriceID].TotalCount = count
	log.Printf("Found %d subscriptions for price ID %s", count, oldPriceID)
}
