import { describe, expect, it } from "bun:test";
import {
	extractPropertiesFromObject,
	extractPropertyFromObject,
} from "../extractPropertyFromObject";

describe("extractPropertyFromObject", () => {
	it("returns the value stored at the key", () => {
		const object = { id: 1, name: "alice" };

		const result = extractPropertyFromObject(object, "name");

		expect(result).toBe("alice");
	});

	it("deletes the extracted key from the object", () => {
		const object = { id: 1, name: "alice" };

		extractPropertyFromObject(object, "name");

		expect("name" in object).toBe(false);
		expect(object).toEqual({ id: 1 });
	});

	it("returns undefined for a key whose value is undefined", () => {
		const object: { a?: number } = { a: undefined };

		const result = extractPropertyFromObject(object, "a");

		expect(result).toBeUndefined();
		expect("a" in object).toBe(false);
	});

	it("throws an error if the key does not exist", () => {
		const object = { id: 1 };

		expect(() =>
			extractPropertyFromObject(object, "name" as keyof typeof object),
		).toThrowError();
	});

	it.skip("fails to compile if the key does not exist without type overrides", () => {
		// This test is skipped on purpose, it is merely a compile-time check
		const object = { id: 1 };

		// Remove the line below to see the error
		// @ts-expect-error - This should fail to compile, it's expected
		extractPropertyFromObject(object, "name");
	});
});

describe("extractPropertiesFromObject", () => {
	it("returns the requested keys and their values under `result` with no errors and deletes the keys from the object", () => {
		const object = { id: 1, name: "alice", role: "admin" };

		const { result, errors } = extractPropertiesFromObject(object, ["name", "role"]);

		expect(result).toEqual({ name: "alice", role: "admin" });
		expect(errors).toEqual([]);
		expect(object).toEqual({ id: 1 });
	});

	it("collects an error per missing key instead of throwing", () => {
		const object = { id: 1, name: "alice" };

		const { result, errors } = extractPropertiesFromObject(object, [
			"name",
			"role",
			"age",
		] as (keyof typeof object)[]);

		expect(result).toEqual({ name: "alice" });
		expect(errors).toEqual(["Key \"role\" does not exist in the object", "Key \"age\" does not exist in the object"]);
		expect(object).toEqual({ id: 1 });
	});

	it("still deletes the present keys when others are missing", () => {
		const object = { id: 1, name: "alice" };

		extractPropertiesFromObject(object, [
			"name",
			"role",
		] as (keyof typeof object)[]);

		expect(object).toEqual({ id: 1 });
	});

	it("returns a fresh `result` object, not a reference to the input", () => {
		const object = { id: 1, name: "alice" };

		const { result } = extractPropertiesFromObject(object, ["name"]);

		expect(result).not.toBe(object as unknown);
	});

	it("returns an empty result with no errors when no keys are requested", () => {
		const object = { id: 1 };

		const { result, errors } = extractPropertiesFromObject(object, []);

		expect(result).toEqual({});
		expect(errors).toEqual([]);
		expect(object).toEqual({ id: 1 });
	});
});
