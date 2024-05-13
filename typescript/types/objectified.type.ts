export type Objectified<T> = {
	[K in keyof T]: T[K];
};
