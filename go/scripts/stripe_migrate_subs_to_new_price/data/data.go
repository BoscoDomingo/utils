package data

// lookupKeyMapping maps the old lookup keys to the new lookup keys
// Because Stripe doesn't support updating subscriptions by lookup key, we need to use the price ID instead.
// This means that we need to find each lookup key's price ID, requiring an extra step.
//
// If you already have the price IDs, please use the `priceIDMapping` instead.
var LookupKeyMapping = map[string]string{
	// Fill with your data
}

// priceIDMapping maps the old price IDs to the new price IDs
var PriceIDMapping = map[string]string{
	// Fill with your data
}
