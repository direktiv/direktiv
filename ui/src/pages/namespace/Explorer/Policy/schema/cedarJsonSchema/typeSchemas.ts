import type {
  CedarAnnotations,
  CedarSchemaAttributeType,
  CedarSchemaType,
} from "./types";
import {
  EntityOrCommonNameSchema,
  EntityTypeNameSchema,
  ExtensionTypeNameSchema,
  PrimitiveTypeNameSchema,
  SchemaTypeReferenceNameSchema,
  isIdentifierPath,
} from "./identifierSchemas";
import { strictRecordWithKeyValidation } from "./utils";
import { z } from "zod";

export const CedarAnnotationsSchema: z.ZodType<CedarAnnotations> = z.record(
  z.string()
);

const schemaTypeMetadataShape = {
  annotations: CedarAnnotationsSchema.optional(),
};

const attributeTypeMetadataShape = {
  ...schemaTypeMetadataShape,
  required: z.boolean().optional(),
};

const createPrimitiveOrCommonTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.union([PrimitiveTypeNameSchema, SchemaTypeReferenceNameSchema]),
      ...metadataShape,
    })
    .strict();

const createSetTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.literal("Set"),
      element: CedarSchemaTypeSchema,
      ...metadataShape,
    })
    .strict();

const createEntityReferenceTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.literal("Entity"),
      name: EntityTypeNameSchema,
      ...metadataShape,
    })
    .strict();

const createRecordTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.literal("Record"),
      attributes: strictRecordWithKeyValidation(
        CedarSchemaAttributeTypeSchema,
        (key) => isIdentifierPath(key),
        "Record attribute names must be valid Cedar identifiers"
      ),
      ...metadataShape,
    })
    .strict();

const createExtensionTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.literal("Extension"),
      name: ExtensionTypeNameSchema,
      ...metadataShape,
    })
    .strict();

const createEntityOrCommonTypeSchema = (metadataShape: z.ZodRawShape) =>
  z
    .object({
      type: z.literal("EntityOrCommon"),
      name: EntityOrCommonNameSchema,
      ...metadataShape,
    })
    .strict();

// `shape`, `tags`, `context`, and common types all use this grammar.
export const CedarSchemaTypeSchema: z.ZodType<CedarSchemaType> = z.lazy(
  () => CedarSchemaTypeUnion
);

// Record attributes reuse the same grammar and add the JSON-only `required` flag.
const CedarSchemaAttributeTypeSchema: z.ZodType<CedarSchemaAttributeType> =
  z.lazy(() => CedarSchemaAttributeTypeUnion);

const CedarPrimitiveOrCommonTypeSchema = createPrimitiveOrCommonTypeSchema(
  schemaTypeMetadataShape
);

const CedarSetTypeSchema = createSetTypeSchema(schemaTypeMetadataShape);

const CedarEntityReferenceTypeSchema = createEntityReferenceTypeSchema(
  schemaTypeMetadataShape
);

const CedarRecordTypeSchema = createRecordTypeSchema(schemaTypeMetadataShape);

const CedarExtensionTypeSchema = createExtensionTypeSchema(
  schemaTypeMetadataShape
);

const CedarEntityOrCommonTypeSchema = createEntityOrCommonTypeSchema(
  schemaTypeMetadataShape
);

const CedarAttributePrimitiveOrCommonTypeSchema =
  createPrimitiveOrCommonTypeSchema(attributeTypeMetadataShape);

const CedarAttributeSetTypeSchema = createSetTypeSchema(
  attributeTypeMetadataShape
);

const CedarAttributeEntityReferenceTypeSchema = createEntityReferenceTypeSchema(
  attributeTypeMetadataShape
);

const CedarAttributeRecordTypeSchema = createRecordTypeSchema(
  attributeTypeMetadataShape
);

const CedarAttributeExtensionTypeSchema = createExtensionTypeSchema(
  attributeTypeMetadataShape
);

const CedarAttributeEntityOrCommonTypeSchema = createEntityOrCommonTypeSchema(
  attributeTypeMetadataShape
);

const CedarSchemaTypeUnion = z.union([
  CedarPrimitiveOrCommonTypeSchema,
  CedarSetTypeSchema,
  CedarEntityReferenceTypeSchema,
  CedarRecordTypeSchema,
  CedarExtensionTypeSchema,
  CedarEntityOrCommonTypeSchema,
]);

const CedarSchemaAttributeTypeUnion = z.union([
  CedarAttributePrimitiveOrCommonTypeSchema,
  CedarAttributeSetTypeSchema,
  CedarAttributeEntityReferenceTypeSchema,
  CedarAttributeRecordTypeSchema,
  CedarAttributeExtensionTypeSchema,
  CedarAttributeEntityOrCommonTypeSchema,
]);
