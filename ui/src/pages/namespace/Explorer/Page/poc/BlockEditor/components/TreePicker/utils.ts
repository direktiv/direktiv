import z from "zod";

export type Tree = {
  [key: string]: Tree | unknown;
};

// please keep the schema in sync with the type definition above
const TreeSchema: z.ZodType<Tree> = z.lazy(() =>
  z.record(z.union([TreeSchema, z.unknown()]))
);

const isTree = (value: unknown): value is Tree =>
  TreeSchema.safeParse(value).success;

const getSubtree = (tree: Tree, path: string[]): Tree | null =>
  path.reduce<Tree | null>((current, segment) => {
    if (current === null) return null;
    const next = current[segment];
    return isTree(next) ? next : null;
  }, tree);

export const getSublist = (tree: Tree, path: string[]): string[] | null => {
  const subtree = getSubtree(tree, path);
  if (subtree === null) {
    return null;
  }
  return Object.keys(subtree);
};
