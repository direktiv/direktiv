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

const getSubtree = (tree: Tree, path: string[]): Tree =>
  path.reduce<Tree>((current, segment) => {
    const next = current[segment];
    return isTree(next) ? next : current;
  }, tree);

export const getSublist = (tree: Tree, path: string[]): string[] => {
  const subtree = getSubtree(tree, path);
  if (typeof subtree === "string" || typeof subtree === "undefined") {
    return [];
  }
  return Object.keys(subtree);
};
