/**
 * Returns a pseudo-random floating-point number in the half-open interval `[min, max)`.
 *
 * @param min inclusive lower bound
 * @param max exclusive upper bound
 * @returns a random float `n` where `min <= n < max`
 */
export function getRandomNumber(min: number, max: number): number {
	return Math.random() * (max - min) + min;
}

/**
 * Returns a pseudo-random integer in the half-open interval `[ceil(min), floor(max))`.
 *
 * The lower bound is inclusive and the upper bound is exclusive.
 *
 * @param min inclusive lower bound (rounded up)
 * @param max exclusive upper bound (rounded down)
 * @returns a random integer `n` where `ceil(min) <= n < floor(max)`
 */
export function getRandomInt(min: number, max: number): number {
	const localMin = Math.ceil(min);
	const localMax = Math.floor(max);
	return Math.floor(Math.random() * (localMax - localMin) + localMin); // The maximum is exclusive and the minimum is inclusive
}
