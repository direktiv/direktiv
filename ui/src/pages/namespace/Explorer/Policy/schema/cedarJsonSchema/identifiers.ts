import {
  type CedarActionEntityTypeName,
  cedarPrimitiveTypeNames,
} from "./types";
import { z } from "zod";

const cedarSchemaTypeKeywords = [
  ...cedarPrimitiveTypeNames,
  "Set",
  "Entity",
  "Record",
  "Extension",
  "EntityOrCommon",
] as const;

// Reserved Cedar type names cannot be reused as user-defined type identifiers.
// Example: `type String = ...` is invalid because `String` is built in.
const reservedCedarTypeNames = new Set(["Bool", ...cedarSchemaTypeKeywords]);

// starts with `_` or a letter, followed by letters, digits, or underscores.
const cedarIdentifierSegmentPattern = /^[_a-zA-Z][_a-zA-Z0-9]*$/;

const usesReservedCedarNamespace = (value: string) =>
  value.startsWith("__cedar::");

// Validates the shared Cedar identifier-path grammar used in namespaces, entity
// types, action entity types, and type references. The optional flags let
// each caller enable the few Cedar-specific exceptions that vary by context,
// such as the empty namespace, reserved `__cedar` references, or reserved final
// segments like `Action`.
export const isIdentifierPath = (
  value: string,
  options: {
    allowEmpty?: boolean;
    allowReservedCedarNamespace?: boolean;
    allowReservedFinalSegment?: boolean;
  } = {}
) => {
  const {
    allowEmpty = false,
    allowReservedCedarNamespace = false,
    allowReservedFinalSegment = false,
  } = options;

  if (value === "") {
    return allowEmpty;
  }

  const segments = value.split("::");

  if (
    segments.some((segment) => !cedarIdentifierSegmentPattern.test(segment))
  ) {
    return false;
  }

  if (
    !allowReservedCedarNamespace &&
    segments.some((segment) => segment === "__cedar")
  ) {
    return false;
  }

  const finalSegment = segments.at(-1);

  if (!finalSegment) {
    return false;
  }

  return allowReservedFinalSegment || !reservedCedarTypeNames.has(finalSegment);
};

const createIdentifierPathSchema = (
  validator: (value: string) => boolean,
  message: string
) => z.string().refine(validator, { message });

export const isNamespaceName = (value: string) =>
  isIdentifierPath(value, { allowEmpty: true });

export const isEntityTypeName = (value: string) => isIdentifierPath(value);

export const isActionName = (value: string) => isIdentifierPath(value);

export const isEntityOrCommonName = (value: string) =>
  isIdentifierPath(value, {
    allowReservedCedarNamespace: usesReservedCedarNamespace(value),
    allowReservedFinalSegment: true,
  });

export const isExtensionTypeName = (value: string) =>
  isIdentifierPath(value, {
    allowReservedCedarNamespace: usesReservedCedarNamespace(value),
    allowReservedFinalSegment: true,
  });

const isActionEntityTypePath = (value: string) =>
  isIdentifierPath(value, { allowReservedFinalSegment: true });

export const isActionEntityTypeName = (value: string) =>
  isActionEntityTypePath(value) && value.split("::").at(-1) === "Action";

export const isSchemaTypeReferenceName = (value: string) =>
  !cedarSchemaTypeKeywords.includes(
    value as (typeof cedarSchemaTypeKeywords)[number]
  ) &&
  isIdentifierPath(value, {
    allowReservedCedarNamespace: usesReservedCedarNamespace(value),
  });

// Entity type references appear in places like:
// Cedar: `entity User in Group;`
// JSON:  `{ "memberOfTypes": ["Group"] }`
export const EntityTypeNameSchema = createIdentifierPathSchema(
  isEntityTypeName,
  "Entity type names must be valid Cedar identifier paths"
);

// `EntityOrCommon` follows Cedar's name resolution rules:
// common type > entity type > primitive/extension type.
export const EntityOrCommonNameSchema = createIdentifierPathSchema(
  isEntityOrCommonName,
  "Entity or common type references must be valid Cedar identifier paths"
);

export const ExtensionTypeNameSchema = createIdentifierPathSchema(
  isExtensionTypeName,
  "Extension type names must be valid Cedar identifier paths"
);

// Action groups are always action entities.
// Cedar: `action View in [ReadOnly] ...`
// JSON:  `{ "memberOf": [{ "id": "ReadOnly", "type": "MyNS::Action" }] }`
export const ActionEntityTypeNameSchema = z
  .string()
  .refine(isActionEntityTypePath, {
    message: "Action entity types must be valid Cedar identifier paths",
  })
  .refine((value) => isActionEntityTypeName(value), {
    message: "Action entity types must end with 'Action'",
  }) as z.ZodType<CedarActionEntityTypeName>;

export const PrimitiveTypeNameSchema = z.enum(cedarPrimitiveTypeNames);

// Common type references are encoded as `{ type: "MyCommonType" }`, so this excludes
// the built-in discriminators like `Record`, `Set`, `Entity`, and primitive names.
export const SchemaTypeReferenceNameSchema = createIdentifierPathSchema(
  isSchemaTypeReferenceName,
  "Type references must be valid Cedar type names"
);
