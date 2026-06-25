import { describe, expect, it } from "bun:test";
import { addToSet, union } from "../set.utils";

describe("union", () => {
	it("combines the elements of both sets", () => {
		const result = union(new Set([1, 2]), new Set([2, 3]));
		expect([...result].sort()).toEqual([1, 2, 3]);
	});

	it("deduplicates overlapping elements", () => {
		const result = union(new Set(["a", "b"]), new Set(["b", "b"]));
		expect(result.size).toBe(2);
		expect(result.has("a")).toBe(true);
		expect(result.has("b")).toBe(true);
	});

	it("returns a new set without mutating the inputs", () => {
		const set1 = new Set([1]);
		const set2 = new Set([2]);
		const result = union(set1, set2);

		expect(result).not.toBe(set1);
		expect(result).not.toBe(set2);
		expect([...set1]).toEqual([1]);
		expect([...set2]).toEqual([2]);
	});

	it("handles empty sets", () => {
		expect(union(new Set<number>(), new Set<number>()).size).toBe(0);
		expect([...union(new Set([1]), new Set<number>())]).toEqual([1]);
	});
});

describe("addToSet", () => {
	it("adds every element of the source set into the target set", () => {
		const target = new Set([1, 2]);
		addToSet(target, new Set([3, 4]));
		expect([...target].sort()).toEqual([1, 2, 3, 4]);
	});

	it("mutates the original set in place and returns nothing", () => {
		const target = new Set([1]);
		const result = addToSet(target, new Set([2]));

		expect(result).toBeUndefined();
		expect([...target].sort()).toEqual([1, 2]);
	});

	it("leaves the source set unchanged", () => {
		const source = new Set([2, 3]);
		addToSet(new Set([1]), source);
		expect([...source].sort()).toEqual([2, 3]);
	});

	it("ignores elements already present", () => {
		const target = new Set([1, 2]);
		addToSet(target, new Set([2, 3]));
		expect(target.size).toBe(3);
	});
});
