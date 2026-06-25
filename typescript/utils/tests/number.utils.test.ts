import { describe, expect, it } from "bun:test";
import {
	isCommaDecimalFormattedNumber,
	parseCommaDecimalFormattedNumber,
} from "../number.utils";

describe("isCommaDecimalFormattedNumber", () => {
	it("accepts a single comma used as a decimal separator", () => {
		expect(isCommaDecimalFormattedNumber("1,00")).toBe(true);
		expect(isCommaDecimalFormattedNumber("1,234")).toBe(true);
	});

	it("accepts dots used as thousand separators", () => {
		expect(isCommaDecimalFormattedNumber("10.000.000")).toBe(true);
		expect(isCommaDecimalFormattedNumber("10.000,01")).toBe(true);
	});

	it("accepts a single dot only when followed by exactly 3 digits", () => {
		expect(isCommaDecimalFormattedNumber("1.000")).toBe(true);
		expect(isCommaDecimalFormattedNumber("1.234")).toBe(true);
		expect(isCommaDecimalFormattedNumber("1.00")).toBe(false);
		expect(isCommaDecimalFormattedNumber("1.0000")).toBe(false);
		expect(isCommaDecimalFormattedNumber("1.2")).toBe(false);
	});

	it("rejects English decimal and plain integer formats", () => {
		expect(isCommaDecimalFormattedNumber("1234")).toBe(false);
	});

	it("rejects strings with multiple commas", () => {
		expect(isCommaDecimalFormattedNumber("1,2,3")).toBe(false);
	});

	it("rejects empty, blank and non-numeric strings", () => {
		expect(isCommaDecimalFormattedNumber("")).toBe(false);
		expect(isCommaDecimalFormattedNumber("   ")).toBe(false);
		expect(isCommaDecimalFormattedNumber("12a")).toBe(false);
	});

	it("ignores surrounding whitespace", () => {
		expect(isCommaDecimalFormattedNumber("  1,00  ")).toBe(true);
		expect(isCommaDecimalFormattedNumber("  1.000  ")).toBe(true);
	});
});

describe("parseCommaDecimalFormattedNumber", () => {
	it("treats the comma as the decimal separator", () => {
		expect(parseCommaDecimalFormattedNumber("10,0")).toBe(10);
	});

	it("strips dots used as thousand separators", () => {
		expect(parseCommaDecimalFormattedNumber("10.0")).toBe(100);
		expect(parseCommaDecimalFormattedNumber("10.000")).toBe(10000);
	});

	it("strips every thousand separator, not just the first", () => {
		expect(parseCommaDecimalFormattedNumber("10.000.000")).toBe(10000000);
	});

	it("handles combined thousand and decimal separators", () => {
		expect(parseCommaDecimalFormattedNumber("10.000,01")).toBe(10000.01);
	});
});
