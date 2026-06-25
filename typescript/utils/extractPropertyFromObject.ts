/**
 * Removes `key` from `object` and returns the value that was stored there.
 *
 * Mutates `object` by deleting the property.
 *
 * @param object object to extract the property from (mutated)
 * @param key key of the property to remove and return
 * @returns the value previously held at `object[key]`
 * @throws {Error} if `key` is not present in `object`
 */
export function extractPropertyFromObject<
	T extends Record<string, unknown>,
	K extends keyof T,
>(object: T, key: K): T[K] {
	if (!(key in object)) {
		throw new Error(`Key "${String(key)}" does not exist in the object`);
	}

	const propertyValue = object[key];
	delete object[key];
	return propertyValue;
}

/**
 * Removes the given `keys` from `object` and returns them as a new record.
 *
 * Warning: Mutates `object` by deleting each listed property.
 *
 * If a key does not exist in the object, it is ignored and an error is added to the errors array.
 * However, the other keys are still extracted and the operation is not interrupted.
 *
 * @param object object to extract the properties from (mutated)
 * @param keys keys of the properties to remove and return
 * @returns `{ result, errors }` where `result` is a new record mapping each found key to its value and `errors` holds one message per missing key
 */
export function extractPropertiesFromObject<
	T extends Record<string, unknown>,
	K extends keyof T,
>(object: T, keys: K[]): { result: Record<K, T[K]>; errors: string[] } {
	const values = {} as Record<K, T[K]>;
	const errors: string[] = [];

	for (const key of keys) {
		if (!(key in object)) {
			errors.push(`Key "${String(key)}" does not exist in the object`);
			continue;
		}

		values[key] = object[key];
		delete object[key];
	}

	return { result: values, errors };
}
