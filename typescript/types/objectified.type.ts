/**
 * Utility type that converts a type to an object.
 *
 * @example
 * type Objectified = Objectified<{ a: string; b: string }>;
 * // Valid: { a: "value", b: "value" }
 */
export type Objectified<T> = { [K in keyof T]: T[K] };
