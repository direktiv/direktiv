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

// Reserved Cedar type names cannot be reused as user defined type identifiers.
// You can e.g. use "type": "PersonType" to point to a common type but not
// "type": "String"
const reservedCedarTypeNames = new Set(["Bool", ...cedarSchemaTypeKeywords]);

// starts with `_` or a letter, followed by letters, digits, or underscores.
const cedarIdentifierSegmentPattern = /^[_a-zA-Z][_a-zA-Z0-9]*$/;

const containsReservedCedarNamespace = (value: string) =>
  value.startsWith("__cedar::");

// Validates the shared Cedar identifier path grammar used in namespaces, entity
// types, action entity types, and type references. The optional flags let
// each caller enable the few Cedar specific exceptions that vary by context,
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
    allowReservedCedarNamespace: containsReservedCedarNamespace(value),
    allowReservedFinalSegment: true,
  });

export const isExtensionTypeName = (value: string) =>
  isIdentifierPath(value, {
    allowReservedCedarNamespace: containsReservedCedarNamespace(value),
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
    allowReservedCedarNamespace: containsReservedCedarNamespace(value),
  });

// Entity type names identify Cedar entity types such as `User` or
// `Namespace::Group`. They are used anywhere the schema names an entity type,
// for example in `{ "memberOfTypes": ["Group"] }` or `"User": { "shape": ... }`.
export const EntityTypeNameSchema = createIdentifierPathSchema(
  isEntityTypeName,
  "Entity type names must be valid Cedar identifier paths"
);

// `EntityOrCommon` references use Cedar's lookup order for type names:
// common type first, then entity type, then primitive or extension type.
// For example, `{ "type": "Address" }` resolves to a common type if one is
// defined, otherwise `{ "type": "User" }` can resolve to an entity type, and
// `{ "type": "String" }` still remains valid as a primitive type reference.
export const EntityOrCommonNameSchema = createIdentifierPathSchema(
  isEntityOrCommonName,
  "Entity or common type references must be valid Cedar identifier paths"
);

// Extension type names refer to Cedar extension types, including reserved
// `__cedar` names when Cedar allows them, for example `ipaddr` or
// `__cedar::Decimal`.
export const ExtensionTypeNameSchema = createIdentifierPathSchema(
  isExtensionTypeName,
  "Extension type names must be valid Cedar identifier paths"
);

// Action entity types are the entity types used for actions themselves. In
// Cedar they must end with `Action`, for example `Action` or `Namespace::Action`.
export const ActionEntityTypeNameSchema = z
  .string()
  .refine(isActionEntityTypePath, {
    message: "Action entity types must be valid Cedar identifier paths",
  })
  .refine((value) => isActionEntityTypeName(value), {
    message: "Action entity types must end with 'Action'",
  }) as z.ZodType<CedarActionEntityTypeName>;

export const PrimitiveTypeNameSchema = z.enum(cedarPrimitiveTypeNames);

// Type references are the names used in `{ "type": "..." }`, such as
// `{ "type": "MyCommonType" }` or `{ "type": "Namespace::Profile" }`.
// Built-in Cedar keywords like `String`, `Set`, `Record`, and `Entity`
// are not valid here.
export const SchemaTypeReferenceNameSchema = createIdentifierPathSchema(
  isSchemaTypeReferenceName,
  "Type references must be valid Cedar type names"
);
