import {
  ActionEntityTypeNameSchema,
  EntityTypeNameSchema,
  isActionName,
  isEntityTypeName,
} from "./identifierSchemas";
import type {
  CedarActionDeclaration,
  CedarActionGroupReference,
  CedarEntityDefinition,
  CedarNamespaceDefinition,
  CedarSchemaType,
} from "./types";
import { CedarAnnotationsSchema, CedarSchemaTypeSchema } from "./typeSchemas";
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

// An entity definition is either:
// - a structural entity with optional membership, shape, tags, and annotations
// - an enum entity whose allowed entity IDs are listed explicitly
const CedarEntityDefinitionSchema: z.ZodType<CedarEntityDefinition> = z.union([
  CedarStructuralEntityDefinitionSchema,
  CedarEnumEntityDefinitionSchema,
]);

// Each `memberOf` entry points to an action group. The optional `type` lets the
// reference use a namespaced action entity type instead of the default `Action`.
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
