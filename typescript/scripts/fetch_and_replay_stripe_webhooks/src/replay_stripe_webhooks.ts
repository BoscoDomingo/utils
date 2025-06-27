const SIGNING_SECRET = Bun.env.SIGNING_SECRET!;
const ENDPOINT_URL = Bun.env.ENDPOINT_URL!;
const INPUT_FILE = "./data/failed_payloads.json";

async function generateStripeSignature(
	payload: string,
	secret: string,
): Promise<string> {
	const timestamp = Math.floor(Date.now() / 1000);
	const signedPayload = `${timestamp}.${payload}`;

	// Create HMAC using Web Crypto API
	const key = await crypto.subtle.importKey(
		"raw",
		new TextEncoder().encode(secret),
		{ name: "HMAC", hash: "SHA-256" },
		false,
		["sign"],
	);

	const signature = await crypto.subtle.sign(
		"HMAC",
		key,
		new TextEncoder().encode(signedPayload),
	);

	const signatureHex = Array.from(new Uint8Array(signature))
		.map((b) => b.toString(16).padStart(2, "0"))
		.join("");

	return `t=${timestamp},v1=${signatureHex}`;
}

const raw = await Bun.file(INPUT_FILE).text();
const events = JSON.parse(raw);

// Track successfully replayed events to avoid duplicates
const replayedEvents = new Set<string>();
let skippedCount = 0;
let successCount = 0;
let failedCount = 0;

for (const event of events) {
	// Skip if we've already successfully replayed this event
	if (replayedEvents.has(event.id)) {
		skippedCount++;
		console.log(`⏭️  Skipping already replayed event ${event.id}`);
		continue;
	}

	const payload = JSON.stringify(event);
	const signature = await generateStripeSignature(payload, SIGNING_SECRET);

	const res = await fetch(ENDPOINT_URL, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			"Stripe-Signature": signature,
		},
		body: payload,
	});

	if (res.ok) {
		replayedEvents.add(event.id);
		successCount++;
		console.log(`✔️  Replayed event ${event.id}`);
	} else {
		failedCount++;
		const text = await res.text();
		console.error(`❌ Failed event ${event.id}: ${res.status} ${text}`);
	}
}

console.log(`\n📊 Summary:`);
console.log(`   ✔️  Successfully replayed: ${successCount} events`);
console.log(`   ❌ Failed: ${failedCount} events`);
console.log(`   ⏭️  Skipped (duplicates): ${skippedCount} events`);
console.log(`   📝 Total processed: ${events.length} events`);

// Make this file a module
export {};
