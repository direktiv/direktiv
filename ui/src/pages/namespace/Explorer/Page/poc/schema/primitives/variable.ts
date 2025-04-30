import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;

export const supportedVariableNamespaces = ["query"] as const;

export const VariableNamespaceSchema = z.enum(supportedVariableNamespaces);

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
