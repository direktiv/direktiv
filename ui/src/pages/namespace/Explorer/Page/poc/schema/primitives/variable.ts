import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;

export const asynchronousVariableNamespaces = ["query"] as const;
export const synchronousVariableNamespaces = ["loop"] as const;

const AsynchronousVariableNamespacesSchema = z.enum(
  asynchronousVariableNamespaces
);

const SynchronousVariableNamespacesSchema = z.enum(
  synchronousVariableNamespaces
);

export type SynchronousVariableNamespace = z.infer<
  typeof SynchronousVariableNamespacesSchema
>;

export const VariableNamespaceSchema = z.union([
  AsynchronousVariableNamespacesSchema,
  SynchronousVariableNamespacesSchema,
]);

export type VariableNamespace = z.infer<typeof VariableNamespaceSchema>;

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
