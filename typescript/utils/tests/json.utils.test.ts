import { describe, expect, it } from "bun:test";
import { jsonStringify } from "../json.utils";

describe("jsonStringify", () => {
	it("serializes plain values like JSON.stringify", () => {
		const value = { a: 1, b: ["x", true, null] };
		expect(jsonStringify(value)).toBe(JSON.stringify(value));
	});

	it("serializes a Set as an array of its values", () => {
		expect(jsonStringify(new Set([1, 2, 3]))).toBe("[1,2,3]");
	});

	it("serializes a Map as an array of [key, value] entries", () => {
		expect(jsonStringify(new Map([["a", 1]]))).toBe('[["a",1]]');
	});

	it("serializes nested Set and Map values", () => {
		const value = { items: new Set(["a", "b"]), counts: new Map([["a", 2]]) };
		expect(jsonStringify(value)).toBe('{"items":["a","b"],"counts":[["a",2]]}');
	});

	it("applies the space argument for pretty-printing", () => {
		expect(jsonStringify({ a: 1 }, 2)).toBe('{\n  "a": 1\n}');
	});

	it("returns an empty object if the value is not serializable", () => {
		expect(jsonStringify(new Error("test"))).toBe("{}");
	});

	it("throws an error if the value is not serializable", () => {
		expect(() => jsonStringify(BigInt(1))).toThrowError();
	});

	it("works with empty Set and Map values", () => {
		expect(jsonStringify(new Set())).toBe("[]");
		expect(jsonStringify(new Map())).toBe("[]");
	});
});
