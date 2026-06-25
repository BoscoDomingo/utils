/**
 * `JSON.stringify` replacer that makes `Set` and `Map` values serializable by
 * spreading them into arrays. Both serialize to `{}` by default.
 */
function jsonReplacerFunction(_key: string, value: unknown): unknown {
	// Necessary for Map and Set objects, which are not serializable by default.
	if (value instanceof Set || value instanceof Map) return [...value];
	return value;
}

/**
 * Serializes a value to JSON like `JSON.stringify`, but with support for `Set`
 * and `Map`.
 *
 * A `Set` is emitted as an array of its values and a `Map` as an array of its
 * `[key, value]` entries (both at any depth).
 *
 * @param object value to serialize
 * @param space indentation forwarded to `JSON.stringify` (space count or string)
 * @returns the JSON string representation of `object`
 */
export function jsonStringify(
	object: unknown,
	space?: string | number,
): string {
	return JSON.stringify(object, jsonReplacerFunction, space);
}
