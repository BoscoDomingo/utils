# Stripe Failed Webhook Replay with Bun

A simple tool to **fetch all failed Stripe webhook events** and **replay** them against your backend using only Bun's built-in APIs — no dependencies required.

---

## Setup

1. Install [Bun](https://bun.sh/).

2. Create a `.env` file with:

```env
STRIPE_SECRET=sk_live_or_test_...
SIGNING_SECRET=whsec_...
ENDPOINT_URL=https://your-backend.com/webhook
```

## Usage

1. Fetch all failed webhook events

`bun fetch-failed-webhooks.ts`

This saves failed_payloads.json containing all failed event payloads.
2. Replay failed webhook events

`bun replay-failed-webhooks.ts`

It reads failed_payloads.json, re-signs, and posts each event to your backend endpoint.

## Notes

- Uses Stripe's Events API with `delivery_success=false` to get failed webhook deliveries.

- Replay script mimics Stripe's signature scheme to pass webhook verification.

- Pure Bun — no external libs or packages needed.

- Adjust `failed_payloads.json` path or concurrency in the scripts as needed.