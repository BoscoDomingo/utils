export function extractPropertyFromObject<T extends object, K extends keyof T>(
	object: T,
	key: K,
): T[K] {
	const propertyValue = object[key];
	delete object[key];
	return propertyValue;
}

export function extractPropertiesFromObject<T extends Record<string, unknown>, K extends keyof T>(
	object: T,
	keys: K[],
): Record<K, T[K]> {
	const values = {} as Record<K, T[K]>;
	for (const key of keys) {
		values[key] = object[key];
		delete object[key];
	}
	return values;
}
