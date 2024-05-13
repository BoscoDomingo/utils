export function getRandomNumber(min: number, max: number): number {
	return Math.random() * (max - min) + min;
}

export function getRandomInt(min: number, max: number): number {
	const localMin = Math.ceil(min);
	const localMax = Math.floor(max);
	return Math.floor(Math.random() * (localMax - localMin) + localMin); // The maximum is exclusive and the minimum is inclusive
}
