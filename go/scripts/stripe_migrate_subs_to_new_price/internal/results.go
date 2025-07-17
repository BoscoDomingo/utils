package internal

// MigrationResult represents the result of a single subscription migration
type MigrationResult struct {
	SubscriptionID string
	OldPriceID     string
	NewPriceID     string
	Success        bool
	Error          error
}

// PriceMigrationStats tracks statistics for a specific price migration (multiple subscriptions)
type PriceMigrationStats struct {
	OldPriceID    string
	NewPriceID    string
	TotalCount    int
	SuccessCount  int
	FailureCount  int
	FailedSubsIDs []string
}
