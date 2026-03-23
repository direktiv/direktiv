import {
  ActionEntityTypeNameSchema,
  EntityTypeNameSchema,
  isActionName,
  isEntityTypeName,
} from "./identifiers";
import type {
  CedarActionDeclaration,
  CedarActionGroupReference,
  CedarEntityDefinition,
  CedarNamespaceDefinition,
  CedarSchemaType,
} from "./types";
import { CedarAnnotationsSchema, CedarSchemaTypeSchema } from "./schemaTypes";
import { strictRecordWithKeyValidation } from "./utils";
import { z } from "zod";

const CedarStructuralEntityDefinitionSchema = z
  .object({
    memberOfTypes: z.array(EntityTypeNameSchema).optional(),
    shape: CedarSchemaTypeSchema.optional(),
    tags: CedarSchemaTypeSchema.optional(),
    annotations: CedarAnnotationsSchema.optional(),
  })
  .strict();

const CedarEnumEntityDefinitionSchema = z
  .object({
    enum: z.array(z.string()).min(1),
    annotations: CedarAnnotationsSchema.optional(),
  })
  .strict();

// Entity definitions are either structural declarations with optional shape/tags,
// or enum-style entities whose valid EIDs are listed explicitly.
const CedarEntityDefinitionSchema: z.ZodType<CedarEntityDefinition> = z.union([
  CedarStructuralEntityDefinitionSchema,
  CedarEnumEntityDefinitionSchema,
]);

// Action `memberOf` entries reference action groups, which are themselves action entities.
const CedarActionGroupReferenceSchema: z.ZodType<CedarActionGroupReference> = z
  .object({
    id: z.string(),
    type: ActionEntityTypeNameSchema.optional(),
  })
  .strict();

const CedarActionAppliesToSchema: z.ZodType<
  CedarActionDeclaration["appliesTo"]
> = z
  .object({
    principalTypes: z.array(EntityTypeNameSchema),
    resourceTypes: z.array(EntityTypeNameSchema),
    context: CedarSchemaTypeSchema.optional(),
  })
  .strict();

// Cedar action declarations are centered around the `appliesTo` block.
const CedarActionDeclarationSchema: z.ZodType<CedarActionDeclaration> = z
  .object({
    memberOf: z.array(CedarActionGroupReferenceSchema).optional(),
    appliesTo: CedarActionAppliesToSchema,
    annotations: CedarAnnotationsSchema.optional(),
  })
  .strict();

const CedarCommonTypeSchema: z.ZodType<CedarSchemaType> = CedarSchemaTypeSchema;

// A namespace always groups entity types, actions, and optional common types.
export const CedarNamespaceDefinitionSchema: z.ZodType<CedarNamespaceDefinition> =
  z
    .object({
      entityTypes: strictRecordWithKeyValidation(
        CedarEntityDefinitionSchema,
        isEntityTypeName,
        "Entity type names must be valid Cedar identifiers"
      ),
      actions: strictRecordWithKeyValidation(
        CedarActionDeclarationSchema,
        isActionName,
        "Action names must be valid Cedar identifiers"
      ),
      commonTypes: strictRecordWithKeyValidation(
        CedarCommonTypeSchema,
        isEntityTypeName,
        "Common type names must be valid Cedar identifiers"
      ).optional(),
      annotations: CedarAnnotationsSchema.optional(),
    })
    .strict();
