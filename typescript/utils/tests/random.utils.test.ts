import { afterEach, describe, expect, it, spyOn } from "bun:test";
import { getRandomInt, getRandomNumber } from "../random.utils";

afterEach(() => {
	spyOn(Math, "random").mockRestore();
});

const mockRandom = (value: number): void => {
	spyOn(Math, "random").mockReturnValue(value);
};

describe("getRandomNumber", () => {
	it("returns the lower bound when Math.random returns 0", () => {
		mockRandom(0);
		expect(getRandomNumber(5, 10)).toBe(5);
	});

	it("scales the random value across the range", () => {
		mockRandom(0.5);
		expect(getRandomNumber(0, 10)).toBe(5);
		mockRandom(0.25);
		expect(getRandomNumber(4, 8)).toBe(5);
	});

	it("stays within [min, max) across many real draws", () => {
		for (let i = 0; i < 1000; i++) {
			const n = getRandomNumber(-3, 7);
			expect(n).toBeGreaterThanOrEqual(-3);
			expect(n).toBeLessThan(7);
		}
	});
});

describe("getRandomInt", () => {
	it("returns the (rounded-up) lower bound when Math.random returns 0", () => {
		mockRandom(0);
		expect(getRandomInt(1, 5)).toBe(1);
		expect(getRandomInt(1.2, 5)).toBe(2);
	});

	it("never reaches the upper bound (exclusive)", () => {
		// Largest value Math.random can return is just below 1.
		mockRandom(0.999999);
		expect(getRandomInt(1, 5)).toBe(4);
	});

	it("returns integers within [ceil(min), floor(max)) across many real draws", () => {
		for (let i = 0; i < 1000; i++) {
			const n = getRandomInt(2, 6);
			expect(Number.isInteger(n)).toBe(true);
			expect(n).toBeGreaterThanOrEqual(2);
			expect(n).toBeLessThan(6);
		}
	});
});
