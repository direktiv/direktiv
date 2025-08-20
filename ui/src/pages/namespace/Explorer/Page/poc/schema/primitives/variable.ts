import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;

const GlobalVariableNamespaces = ["query", "loop"] as const;
const LocalVariableNamespace = "this" as const;

export type GlobalVariableNamespace = (typeof GlobalVariableNamespaces)[number];
export type LocalVariableNamespace = typeof LocalVariableNamespace;
type VariableNamespace = GlobalVariableNamespace | LocalVariableNamespace;

export const VariableNamespaceSchema = z.enum([
  ...GlobalVariableNamespaces,
  LocalVariableNamespace,
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
