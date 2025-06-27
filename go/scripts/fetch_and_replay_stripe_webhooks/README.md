# Stripe Failed Webhook Replay with Go

A **simple, single-file tool** to fetch all failed Stripe webhook events and replay them against your backend using only Go's standard library — no external dependencies required.

---

## Setup

1. Install [Go](https://golang.org/dl/) (version 1.24 or later).

2. Create a `.env` file or set environment variables:

```bash
# Option 1: Create a .env file (recommended)
cp .env.example .env
# Edit .env with your actual values

# Option 2: Export environment variables
export STRIPE_SECRET=sk_live_or_test_...
export SIGNING_SECRET=whsec_...
export ENDPOINT_URL=https://your-backend.com/webhook
```

## Usage

**Note**: The tool will automatically create a `data/` directory to store JSON files.

### Option 1: Fetch and Replay in One Command (Recommended)

```bash
cd go/scripts/fetch_and_replay_stripe_webhooks
go run src/main.go fetch-and-replay
```

This will fetch all failed events and immediately replay them in sequence.

### Option 2: Separate Commands

1. **Fetch all failed webhook events**

```bash
go run src/main.go fetch
```

This saves `data/failed_payloads.json` containing all failed event payloads.

2. **Replay failed webhook events**

```bash
go run src/main.go replay
```

It reads `data/failed_payloads.json`, re-signs, and posts each event to your backend endpoint with deduplication.

## Building a Binary

You can also build a standalone binary:

```bash
# Build the tool
go build -o bin/stripe-webhook-tool src/main.go

# Use the binary
./bin/stripe-webhook-tool fetch-and-replay  # Recommended: do both at once
./bin/stripe-webhook-tool fetch             # Or fetch only
./bin/stripe-webhook-tool replay            # Or replay only
```

## Features

- **Single file**: Everything in one simple `main.go` file with 0 external dependencies
- **Deduplication**: Automatically skips events that have already been successfully replayed
- **Summary report**: Displays final statistics (success/failed/skipped counts)
- **Error handling**: Proper error messages and exit codes
- **.env file support**: Automatically loads environment variables from `.env` file

## Notes

- Uses Stripe's Events API with `delivery_success=false` to get failed webhook deliveries
- Replay script mimics Stripe's signature scheme to pass webhook verification
- Failed events are not marked as "replayed" so they can be retried in future runs
- Successfully replayed events are tracked to prevent duplicates within the same run
- JSON data is stored in the `data/` directory for better organization