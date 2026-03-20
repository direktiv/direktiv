import {
  ActionEntityTypeSchema,
  CommonTypeReferenceSchema,
  EntityTypeNameSchema,
  ExtensionTypeNameSchema,
  NamespaceNameSchema,
  PrimitiveTypeNameSchema,
  TypeReferenceSchema,
  isIdentifierPath,
} from "./identifiers";
import {
  type CedarActionDefinitionInput,
  type CedarActionGroupMembershipInput,
  type CedarAnnotations,
  type CedarEntityTypeDefinitionInput,
  type CedarJsonSchemaRecord,
  type CedarNamespaceDefinitionInput,
  type CedarRecordAttributeInput,
  type CedarRootTypeInput,
} from "./types";
import { recordWithValidatedKeys } from "./utils";
import { z } from "zod";

export type { CedarJsonSchemaRecord as CedarJsonSchemaInputType } from "./types";

const AnnotationsSchema: z.ZodType<CedarAnnotations> = z.record(z.string());

// Builds the recursive Cedar type grammar used by several JSON schema fields.
// The `allowRequired` switch exists because only record attributes may carry a
// `required` flag; root types such as `shape`, `tags`, or `context` may not.
const buildTypeSchema = (options: { allowRequired: boolean }) => {
  const metadata = {
    annotations: AnnotationsSchema.optional(),
    ...(options.allowRequired ? { required: z.boolean().optional() } : {}),
  };

  const TypeSchema: z.ZodType<CedarRootTypeInput | CedarRecordAttributeInput> =
    z.lazy(() =>
      z.union([
        // Primitive Cedar types and common type aliases both use `{ type: ... }`.
        // Examples:
        // - `{ type: "String" }`
        // - `{ type: "MyCommonType" }`
        z
          .object({
            type: z.union([PrimitiveTypeNameSchema, TypeReferenceSchema]),
            ...metadata,
          })
          .strict(),
        // Cedar: `Set<User>` -> JSON: `{ type: "Set", element: { type: "Entity", name: "User" } }`
        z
          .object({
            type: z.literal("Set"),
            element: RootTypeSchema,
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
            attributes: recordWithValidatedKeys(
              RecordAttributeSchema,
              () => true,
              "Record attribute names must be strings"
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
            name: CommonTypeReferenceSchema,
            ...metadata,
          })
          .strict(),
      ])
    );

  return TypeSchema;
};

const RootTypeSchema = buildTypeSchema({
  allowRequired: false,
}) as z.ZodType<CedarRootTypeInput>;

const RecordAttributeSchema = buildTypeSchema({
  allowRequired: true,
}) as z.ZodType<CedarRecordAttributeInput>;

// Entity definitions are either regular entity declarations with optional shape/tags,
// or enum-style entities whose valid EIDs are listed explicitly.
const EntityTypeDefinitionSchema: z.ZodType<CedarEntityTypeDefinitionInput> =
  z.union([
    z
      .object({
        memberOfTypes: z.array(EntityTypeNameSchema).optional(),
        shape: RootTypeSchema.optional(),
        tags: RootTypeSchema.optional(),
        annotations: AnnotationsSchema.optional(),
      })
      .strict(),
    z
      .object({
        // Cedar: `entity Group enum ["admins", "reviewers"];`
        enum: z.array(z.string()).min(1),
        annotations: AnnotationsSchema.optional(),
      })
      .strict(),
  ]);

// Action `memberOf` entries reference action groups, which are themselves action entities.
const ActionGroupMembershipSchema: z.ZodType<CedarActionGroupMembershipInput> =
  z
    .object({
      id: z.string(),
      type: ActionEntityTypeSchema.optional(),
    })
    .strict();

// In Cedar syntax this corresponds to the `appliesTo` block on an action declaration.
const AppliesToSchema: z.ZodType<CedarActionDefinitionInput["appliesTo"]> = z
  .object({
    principalTypes: z.array(EntityTypeNameSchema),
    resourceTypes: z.array(EntityTypeNameSchema),
    context: RootTypeSchema.optional(),
  })
  .strict();

// Cedar:
// `action View in [ReadOnly] appliesTo { principal: User, resource: Doc, context: {...} };`
const ActionDefinitionSchema: z.ZodType<CedarActionDefinitionInput> = z
  .object({
    memberOf: z.array(ActionGroupMembershipSchema).optional(),
    appliesTo: AppliesToSchema,
    annotations: AnnotationsSchema.optional(),
  })
  .strict();

const CommonTypeDefinitionSchema: z.ZodType<CedarRootTypeInput> =
  RootTypeSchema;

// A namespace groups the three major Cedar schema sections.
// JSON order is not important, but the structure is always:
// `{ entityTypes: ..., actions: ..., commonTypes?: ..., annotations?: ... }`
const NamespaceDefinitionSchema: z.ZodType<CedarNamespaceDefinitionInput> = z
  .object({
    entityTypes: recordWithValidatedKeys(
      EntityTypeDefinitionSchema,
      (key) => isIdentifierPath(key),
      "Entity type names must be valid Cedar identifiers"
    ),
    actions: z.record(ActionDefinitionSchema),
    commonTypes: recordWithValidatedKeys(
      CommonTypeDefinitionSchema,
      (key) => isIdentifierPath(key),
      "Common type names must be valid Cedar identifiers"
    ).optional(),
    annotations: AnnotationsSchema.optional(),
  })
  .strict();

// Top-level Cedar JSON schema:
// {
//   "MyNamespace": { ... },
//   "": { ... } // optional empty namespace
// }
export const CedarJsonSchema: z.ZodType<CedarJsonSchemaRecord> =
  recordWithValidatedKeys(
    NamespaceDefinitionSchema,
    (key) => NamespaceNameSchema.safeParse(key).success,
    "Namespace names must be empty or valid Cedar identifier paths"
  ).refine((value) => Object.keys(value).length > 0, {
    message: "A Cedar schema must declare at least one namespace",
  });
