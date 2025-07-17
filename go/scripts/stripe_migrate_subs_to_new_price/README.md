# Stripe Migrate Subs to New Price

This script migrates all active subscriptions in Stripe from old prices to new prices.

The prices can be referenced by their lookup keys or price IDs.

## Usage

Add the data to the [`data/data.go`](data/data.go) file.

Map each old price to its substitute.

## Build

```bash
go build -o bin/migrate cmd/migrate
```

## Run

`STRIPE_API_KEY` must be set in the environment. You can achieve this using dotenvx or dotenv-cli or `export STRIPE_API_KEY=...` (not recommended, your keys will be exposed in your shell history).

```bash
dotenvx run ./bin/migrate
```