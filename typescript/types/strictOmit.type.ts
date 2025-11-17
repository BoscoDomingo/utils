/**
 * Type-safe version of the built-in {@link Omit} type.
 *
 * @example
 * type Example = StrictOmit<{ a: string; b: string; c: number }, "c">;
 * // Valid: { a: string; b: string }
 *
 * @see https://github.com/microsoft/TypeScript/issues/30825
 */
export type StrictOmit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
