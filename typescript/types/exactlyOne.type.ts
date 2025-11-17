/**
 * Utility type that creates a union where exactly one key from T is required,
 * and all other keys from T are explicitly forbidden (never).
 *
 * @example
 * type OnlyOne = ExactlyOne<{ a: string; b: string }>;
 * // Valid: { a: "value" }
 * // Valid: { b: "value" }
 * // Invalid: { a: "value", b: "value" }
 */
export type ExactlyOne<T, K extends keyof T = keyof T> = K extends keyof T
	? { [P in K]: T[P] } & { [P in Exclude<keyof T, K>]?: never }
	: never;