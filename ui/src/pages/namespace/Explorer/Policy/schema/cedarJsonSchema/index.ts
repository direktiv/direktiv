import { CedarNamespaceDefinitionSchema } from "./definitionSchemas";
import type { CedarSchemaNamespacesInput } from "./types";
import { isNamespaceName } from "./identifiers";
import { strictRecordWithKeyValidation } from "./utils";
import { z } from "zod";

const CedarNamespaceMapSchema = strictRecordWithKeyValidation(
  CedarNamespaceDefinitionSchema,
  isNamespaceName,
  "Namespace names must be empty or valid Cedar identifier paths"
);

// Top-level Cedar JSON schema:
// {
//   "MyNamespace": { ... },
//   "": { ... } // optional empty namespace
// }
export const CedarJsonSchema: z.ZodType<CedarSchemaNamespacesInput> =
  CedarNamespaceMapSchema.refine((value) => Object.keys(value).length > 0, {
    message: "A Cedar schema must declare at least one namespace",
  });
