package data

// lookupKeyMapping maps the old lookup keys to the new lookup keys
// Because Stripe doesn't support finding subscriptions by lookup key, we need to use the price ID instead.
// This means that we need to find each lookup key's price ID, requiring an extra step.
//
// If you already have the price IDs, please use the `priceIDMapping` instead.
var LookupKeyMapping = map[string]string{
	"BASIC-WEEKLY":        "BASIC-WEEKLY_INITIAL-RELEASE",
	"BASIC-MONTHLY":       "BASIC-MONTHLY_INITIAL-RELEASE",
	"BASIC-HALF-YEARLY":   "BASIC-HALF-YEARLY_INITIAL-RELEASE",
	"PREMIUM-WEEKLY":      "PREMIUM-WEEKLY_INITIAL-RELEASE",
	"PREMIUM-MONTHLY":     "PREMIUM-MONTHLY_INITIAL-RELEASE",
	"PREMIUM-HALF-YEARLY": "PREMIUM-HALF-YEARLY_INITIAL-RELEASE",
}

// priceIDMapping maps the old price IDs to the new price IDs
var PriceIDMapping = map[string]string{
	// Fill with your data
}
