/**
 * Returns a new set containing every element from both input sets (their union).
 *
 * Neither input set is mutated.
 *
 * @param set1 first set to combine
 * @param set2 second set to combine
 * @returns a new {@link Set} holding the elements of `set1` and `set2`
 */
export function union<T>(set1: Set<T>, set2: Set<T>): Set<T> {
	return new Set([...set1, ...set2]);
}

/**
 * Adds every element of `setToBeAdded` into `originalSet`, in place.
 *
 * Mutates `originalSet`; `setToBeAdded` is left unchanged.
 *
 * @param originalSet set that receives the new elements (mutated)
 * @param setToBeAdded set whose elements are copied into `originalSet`
 */
export function addToSet<T>(originalSet: Set<T>, setToBeAdded: Set<T>): void {
	for (const item of setToBeAdded) {
		originalSet.add(item);
	}
}
