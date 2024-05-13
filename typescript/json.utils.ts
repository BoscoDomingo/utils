function jsonReplacerFunction(_key: string, value: any): any {
	// Necessary for Map and Set objects, which are not serializable by default.
	if (value instanceof Set || value instanceof Map) return [...value];
	return value;
}

export function jsonStringify(object: unknown, space?: string | number): string {
	return JSON.stringify(object, jsonReplacerFunction, space);
}
