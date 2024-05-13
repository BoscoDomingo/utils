/**
 * Checks if a number (in string) is written in Spanish format.
 *
 * Single-comma instances (e.g. `"1,00"`) and multi-dot instances (e.g. `"1.000.000"`) return `true`
 *
 * Single-dot instances (e.g. `"1.000"`) return `true` only if there are exactly 3 digits after the dot, as it could also represent the decimal separator in English-formatted numbers otherwise.
 *
 * @param number string that represents a Spanish-formatted number (using commas for decimals and dots for thousand separators)
 */
export function isSpanishFormattedNumber(number: string): boolean {
	const trimmedNumber = number.trim();
	if (!trimmedNumber || trimmedNumber.match(/[^0-9,.]/g)) {
		return false;
	}

	const firstCommaIndex = trimmedNumber.indexOf(",");
	const hasCommas = firstCommaIndex > -1;
	const hasMultipleCommas = trimmedNumber.indexOf(",", firstCommaIndex + 1) > -1;

	const firstDotIndex = trimmedNumber.indexOf(".");
	const hasDots = firstDotIndex > -1;
	const hasMultipleDots = trimmedNumber.indexOf(".", firstDotIndex + 1) > -1;

	return (
		!hasMultipleCommas &&
		((hasCommas && !hasDots) || // 1,234
			(hasCommas && firstDotIndex < firstCommaIndex) || //10.000,01
			hasMultipleDots || // 10.000.000
			(!hasCommas && hasDots && number.slice(firstDotIndex + 1).length === 3)) // 1.234
	);
}

/**
 * Converts a Spanish-formatted number (as string) to a number by removing all dots and replacing any commas with dots.
 *
 * Assumes formatting is correct.
 *
 * @example parseSpanishFormattedNumber("10,0") => 10
 * @example parseSpanishFormattedNumber("10.0") => 100
 * @example parseSpanishFormattedNumber("10.000") => 10
 * @example parseSpanishFormattedNumber("10.000,01") => 10000.01
 * @example parseSpanishFormattedNumber("10.000.000") => 10000000
 * @param number string that represents a Spanish-formatted number (using commas for decimals and dots for thousand separators)
 */
export function parseSpanishFormattedNumber(number: string): number {
	return +number.replace(".", "").replace(",", ".");
}
