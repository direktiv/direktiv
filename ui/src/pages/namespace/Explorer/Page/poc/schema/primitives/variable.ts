import { z } from "zod";

export const Variable = z.string().min(1);

export type VariableType = z.infer<typeof Variable>;

export const contextVariableNamespaces = ["query", "loop"] as const;
export const localVariableNamespace = "this" as const;

type ContextVariableNamespace = (typeof contextVariableNamespaces)[number];

export type LocalVariableNamespace = typeof localVariableNamespace;

type VariableNamespace = ContextVariableNamespace | LocalVariableNamespace;

export const VariableNamespaceSchema = z.enum([
  ...contextVariableNamespaces,
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

export type VariableObjectValidated =
  | {
      src: string;
      namespace: ContextVariableNamespace;
      id: string;
      pointer: string;
    }
  | {
      src: string;
      namespace: LocalVariableNamespace;
      id: string;
      pointer?: never;
    };
