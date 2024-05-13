// Source: https://github.com/microsoft/TypeScript/issues/30825
export type StrictOmit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>