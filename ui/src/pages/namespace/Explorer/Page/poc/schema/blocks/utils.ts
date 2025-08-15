export type ExtractUnionFromSet<T> = T extends Set<infer U> ? U : never;
