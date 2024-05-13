export function union<T>(set1: Set<T>, set2: Set<T>): Set<T> {
	return new Set([...set1, ...set2]);
}

export function addToSet<T>(originalSet: Set<T>, setToBeAdded: Set<T>): void {
	for (const item of setToBeAdded) {
		originalSet.add(item);
	}
}
