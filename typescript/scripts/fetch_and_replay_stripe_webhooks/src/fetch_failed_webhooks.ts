const STRIPE_SECRET = Bun.env.STRIPE_SECRET!;
const OUTPUT_FILE = "./data/failed_payloads.json";

const BASE = "https://api.stripe.com/v1";

async function fetchJson(path: string, params: Record<string, string> = {}) {
	const query = new URLSearchParams(params).toString();
	const url = `${BASE}${path}?${query}`;
	const res = await fetch(url, {
		headers: {
			Authorization: `Basic ${btoa(`${STRIPE_SECRET}:`)}`,
		},
	});
	if (!res.ok) throw new Error(`Failed: ${res.status} ${await res.text()}`);
	return res.json();
}

async function fetchFailedEvents() {
	const events: any[] = [];
	let starting_after: string | undefined;

	while (true) {
		const data = await fetchJson("/events", {
			limit: "100",
			delivery_success: "false",
			...(starting_after ? { starting_after } : {}),
		});

		events.push(...data.data);
		if (!data.has_more) break;
		starting_after = data.data.at(-1).id;
	}

	await Bun.write(OUTPUT_FILE, JSON.stringify(events, null, 2));
	console.log(`Saved ${events.length} failed events to ${OUTPUT_FILE}`);
}

await fetchFailedEvents();
