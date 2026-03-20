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

const AnnotationsSchema: z.ZodType<CedarAnnotations> = z.record(z.string());

let RootTypeSchema: z.ZodType<CedarRootTypeInput>;
let RecordAttributeSchema: z.ZodType<CedarRecordAttributeInput>;

const buildTypeSchema = (options: { allowRequired: boolean }) => {
  const metadata = {
    annotations: AnnotationsSchema.optional(),
    ...(options.allowRequired ? { required: z.boolean().optional() } : {}),
  };

  const TypeSchema: z.ZodType<CedarRootTypeInput | CedarRecordAttributeInput> =
    z.lazy(() =>
      z.union([
        z
          .object({
            type: z.union([PrimitiveTypeNameSchema, TypeReferenceSchema]),
            ...metadata,
          })
          .strict(),
        z
          .object({
            type: z.literal("Set"),
            element: RootTypeSchema,
            ...metadata,
          })
          .strict(),
        z
          .object({
            type: z.literal("Entity"),
            name: EntityTypeNameSchema,
            ...metadata,
          })
          .strict(),
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
        z
          .object({
            type: z.literal("Extension"),
            name: ExtensionTypeNameSchema,
            ...metadata,
          })
          .strict(),
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

RootTypeSchema = buildTypeSchema({
  allowRequired: false,
}) as z.ZodType<CedarRootTypeInput>;
RecordAttributeSchema = buildTypeSchema({
  allowRequired: true,
}) as z.ZodType<CedarRecordAttributeInput>;

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
        enum: z.array(z.string()).min(1),
        annotations: AnnotationsSchema.optional(),
      })
      .strict(),
  ]);

const ActionGroupMembershipSchema: z.ZodType<CedarActionGroupMembershipInput> =
  z
    .object({
      id: z.string(),
      type: ActionEntityTypeSchema.optional(),
    })
    .strict();

const AppliesToSchema: z.ZodType<CedarActionDefinitionInput["appliesTo"]> = z
  .object({
    principalTypes: z.array(EntityTypeNameSchema),
    resourceTypes: z.array(EntityTypeNameSchema),
    context: RootTypeSchema.optional(),
  })
  .strict();

const ActionDefinitionSchema: z.ZodType<CedarActionDefinitionInput> = z
  .object({
    memberOf: z.array(ActionGroupMembershipSchema).optional(),
    appliesTo: AppliesToSchema,
    annotations: AnnotationsSchema.optional(),
  })
  .strict();

const CommonTypeDefinitionSchema: z.ZodType<CedarRootTypeInput> =
  RootTypeSchema;

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

export const CedarJsonSchema: z.ZodType<CedarJsonSchemaRecord> =
  recordWithValidatedKeys(
    NamespaceDefinitionSchema,
    (key) => NamespaceNameSchema.safeParse(key).success,
    "Namespace names must be empty or valid Cedar identifier paths"
  ).refine((value) => Object.keys(value).length > 0, {
    message: "A Cedar schema must declare at least one namespace",
  });
