// biome-ignore lint/complexity/noBannedTypes: <explanation>
export type StrictOmit<T, K extends keyof T | (string & {}) | (number & {}) | (symbol | {})> = {
	[P in Exclude<keyof T, K>]: T[P];
};
