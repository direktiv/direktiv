import {
  ActionEntityTypeNameSchema,
  EntityOrCommonNameSchema,
  EntityTypeNameSchema,
  ExtensionTypeNameSchema,
  NamespaceNameSchema,
  PrimitiveTypeNameSchema,
  SchemaTypeReferenceNameSchema,
  isIdentifierPath,
} from "./identifiers";
import {
  type CedarActionDeclarationInput,
  type CedarActionGroupReferenceInput,
  type CedarAnnotations,
  type CedarEntityDefinitionInput,
  type CedarNamespaceDefinitionInput,
  type CedarSchemaAttributeTypeInput,
  type CedarSchemaNamespacesInput,
  type CedarSchemaTypeInput,
} from "./types";
import { strictRecordWithKeyValidation } from "./utils";
import { z } from "zod";

export type { CedarSchemaNamespacesInput as CedarJsonSchemaInputType } from "./types";

const CedarAnnotationsSchema: z.ZodType<CedarAnnotations> = z.record(
  z.string()
);

// Builds the recursive Cedar type grammar used by several JSON schema fields.
// The `allowRequired` switch exists because only record attributes may carry a
// `required` flag; root types such as `shape`, `tags`, or `context` may not.
const createSchemaTypeValidator = (options: { allowRequired: boolean }) => {
  const metadata = {
    annotations: CedarAnnotationsSchema.optional(),
    ...(options.allowRequired ? { required: z.boolean().optional() } : {}),
  };

  const SchemaTypeValidator: z.ZodType<
    CedarSchemaTypeInput | CedarSchemaAttributeTypeInput
  > = z.lazy(() =>
    z.union([
      // Primitive Cedar types and common type aliases both use `{ type: ... }`.
      // Examples:
      // - `{ type: "String" }`
      // - `{ type: "MyCommonType" }`
      z
        .object({
          type: z.union([
            PrimitiveTypeNameSchema,
            SchemaTypeReferenceNameSchema,
          ]),
          ...metadata,
        })
        .strict(),
      // Cedar: `Set<User>` -> JSON: `{ type: "Set", element: { type: "Entity", name: "User" } }`
      z
        .object({
          type: z.literal("Set"),
          element: CedarSchemaTypeSchema,
          ...metadata,
        })
        .strict(),
      // Cedar entity references such as `owner: User` become explicit in JSON.
      z
        .object({
          type: z.literal("Entity"),
          name: EntityTypeNameSchema,
          ...metadata,
        })
        .strict(),
      // Record values are the most nested part of the grammar.
      // Cedar: `{ owner: User, age?: Long }`
      z
        .object({
          type: z.literal("Record"),
          attributes: strictRecordWithKeyValidation(
            CedarSchemaAttributeTypeSchema,
            (key) => isIdentifierPath(key),
            "Record attribute names must be valid Cedar identifiers"
          ),
          ...metadata,
        })
        .strict(),
      // Cedar extension types like `ipaddr` or `decimal`.
      z
        .object({
          type: z.literal("Extension"),
          name: ExtensionTypeNameSchema,
          ...metadata,
        })
        .strict(),
      // This mirrors the Cedar CLI JSON output when a reference could resolve to
      // either a common type or an entity type.
      z
        .object({
          type: z.literal("EntityOrCommon"),
          name: EntityOrCommonNameSchema,
          ...metadata,
        })
        .strict(),
    ])
  );

  return SchemaTypeValidator;
};

const CedarSchemaTypeSchema = createSchemaTypeValidator({
  allowRequired: false,
}) as z.ZodType<CedarSchemaTypeInput>;

const CedarSchemaAttributeTypeSchema = createSchemaTypeValidator({
  allowRequired: true,
}) as z.ZodType<CedarSchemaAttributeTypeInput>;

// Entity definitions are either regular entity declarations with optional shape/tags,
// or enum-style entities whose valid EIDs are listed explicitly.
const CedarEntityDefinitionSchema: z.ZodType<CedarEntityDefinitionInput> =
  z.union([
    z
      .object({
        memberOfTypes: z.array(EntityTypeNameSchema).optional(),
        shape: CedarSchemaTypeSchema.optional(),
        tags: CedarSchemaTypeSchema.optional(),
        annotations: CedarAnnotationsSchema.optional(),
      })
      .strict(),
    z
      .object({
        // Cedar: `entity Group enum ["admins", "reviewers"];`
        enum: z.array(z.string()).min(1),
        annotations: CedarAnnotationsSchema.optional(),
      })
      .strict(),
  ]);

// Action `memberOf` entries reference action groups, which are themselves action entities.
const CedarActionGroupReferenceSchema: z.ZodType<CedarActionGroupReferenceInput> =
  z
    .object({
      id: z.string(),
      type: ActionEntityTypeNameSchema.optional(),
    })
    .strict();

// In Cedar syntax this corresponds to the `appliesTo` block on an action declaration.
const CedarActionAppliesToSchema: z.ZodType<
  CedarActionDeclarationInput["appliesTo"]
> = z
  .object({
    principalTypes: z.array(EntityTypeNameSchema),
    resourceTypes: z.array(EntityTypeNameSchema),
    context: CedarSchemaTypeSchema.optional(),
  })
  .strict();

// Cedar:
// `action View in [ReadOnly] appliesTo { principal: User, resource: Doc, context: {...} };`
const CedarActionDeclarationSchema: z.ZodType<CedarActionDeclarationInput> = z
  .object({
    memberOf: z.array(CedarActionGroupReferenceSchema).optional(),
    appliesTo: CedarActionAppliesToSchema,
    annotations: CedarAnnotationsSchema.optional(),
  })
  .strict();

const CedarCommonTypeSchema: z.ZodType<CedarSchemaTypeInput> =
  CedarSchemaTypeSchema;

// A namespace groups the three major Cedar schema sections.
// JSON order is not important, but the structure is always:
// `{ entityTypes: ..., actions: ..., commonTypes?: ..., annotations?: ... }`
const CedarNamespaceDefinitionSchema: z.ZodType<CedarNamespaceDefinitionInput> =
  z
    .object({
      entityTypes: strictRecordWithKeyValidation(
        CedarEntityDefinitionSchema,
        (key) => isIdentifierPath(key),
        "Entity type names must be valid Cedar identifiers"
      ),
      actions: strictRecordWithKeyValidation(
        CedarActionDeclarationSchema,
        (key) => isIdentifierPath(key),
        "Action names must be valid Cedar identifiers"
      ),
      commonTypes: strictRecordWithKeyValidation(
        CedarCommonTypeSchema,
        (key) => isIdentifierPath(key),
        "Common type names must be valid Cedar identifiers"
      ).optional(),
      annotations: CedarAnnotationsSchema.optional(),
    })
    .strict();

// Top-level Cedar JSON schema:
// {
//   "MyNamespace": { ... },
//   "": { ... } // optional empty namespace
// }
export const CedarJsonSchema: z.ZodType<CedarSchemaNamespacesInput> =
  strictRecordWithKeyValidation(
    CedarNamespaceDefinitionSchema,
    (key) => NamespaceNameSchema.safeParse(key).success,
    "Namespace names must be empty or valid Cedar identifier paths"
  ).refine((value) => Object.keys(value).length > 0, {
    message: "A Cedar schema must declare at least one namespace",
  });
