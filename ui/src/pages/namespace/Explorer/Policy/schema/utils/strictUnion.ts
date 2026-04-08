type UnionKeys<T> = T extends T ? keyof T : never;

type StrictUnionMember<T, TAll> = T extends T
  ? T & Partial<Record<Exclude<UnionKeys<TAll>, keyof T>, never>>
  : never;

export type StrictUnion<T> = StrictUnionMember<T, T>;
