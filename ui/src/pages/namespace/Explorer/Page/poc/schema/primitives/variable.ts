import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;

const globalVariableNamespaces = ["query", "loop"] as const;
const localVariableNamespace = "this" as const;

export type GlobalVariableNamespace = (typeof globalVariableNamespaces)[number];
export type LocalVariableNamespace = typeof localVariableNamespace;
type VariableNamespace = GlobalVariableNamespace | LocalVariableNamespace;

export const VariableNamespaceSchema = z.enum([
  ...globalVariableNamespaces,
  localVariableNamespace,
]);

/**
 * structured representation of a variable string.
 *
 * Example: "query.company-list.data.0.name" will be represented as:
 * {
 *   src: "query.company-list.data.0.name",
 *   namespace: "query",
 *   id: "company-list",
 *   pointer: "data.0.name"
 * }
 */
export type VariableObject = {
  src: string;
  namespace: VariableNamespace | undefined;
  id: string | undefined;
  pointer: string | undefined;
};

export type VariableObjectValidated = {
  src: string;
  namespace: VariableNamespace;
  id: string;
  pointer: string;
};
