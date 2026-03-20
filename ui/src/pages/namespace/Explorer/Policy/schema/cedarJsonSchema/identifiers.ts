import type { CedarActionEntityTypeName } from "./types";
import { z } from "zod";

// Reserved Cedar type names cannot be reused as user-defined type identifiers.
// Example: `type String = ...` is invalid because `String` is built in.
const reservedCedarTypeNames = new Set([
  "Bool",
  "Boolean",
  "Entity",
  "Extension",
  "Long",
  "Record",
  "Set",
  "String",
]);

const cedarIdentifierSegmentPattern = /^[_a-zA-Z][_a-zA-Z0-9]*$/;

// Validates Cedar-style names like `User`, `MyNamespace::User`, or `__cedar::ipaddr`.
// The options let each caller opt into the few exceptions allowed by the Cedar spec.
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

export const NamespaceNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value, { allowEmpty: true }), {
    message: "Namespace names must be empty or valid Cedar identifier paths",
  });

// Entity type references appear in places like:
// Cedar: `entity User in Group;`
// JSON:  `{ "memberOfTypes": ["Group"] }`
export const EntityTypeNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value), {
    message: "Entity type names must be valid Cedar identifier paths",
  });

// `EntityOrCommon` follows Cedar's name resolution rules:
// common type > entity type > primitive/extension type.
export const EntityOrCommonNameSchema = z.string().refine(
  (value) =>
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
      allowReservedFinalSegment: true,
    }),
  {
    message: "Common type references must be valid Cedar identifier paths",
  }
);

export const ExtensionTypeNameSchema = z.string().refine(
  (value) =>
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
      allowReservedFinalSegment: true,
    }),
  {
    message: "Extension type names must be valid Cedar identifier paths",
  }
);

// Action groups are always action entities.
// Cedar: `action View in [ReadOnly] ...`
// JSON:  `{ "memberOf": [{ "id": "ReadOnly", "type": "MyNS::Action" }] }`
export const ActionEntityTypeNameSchema = z
  .string()
  .refine(
    (value) => isIdentifierPath(value, { allowReservedFinalSegment: true }),
    {
      message: "Action entity types must be valid Cedar identifier paths",
    }
  )
  .refine((value) => value.split("::").at(-1) === "Action", {
    message: "Action entity types must end with 'Action'",
  }) as z.ZodType<CedarActionEntityTypeName>;

export const PrimitiveTypeNameSchema = z.enum(["Long", "String", "Boolean"]);

// Common type references are encoded as `{ type: "MyCommonType" }`, so this excludes
// the built-in discriminators like `Record`, `Set`, `Entity`, and primitive names.
export const SchemaTypeReferenceNameSchema = z.string().refine(
  (value) =>
    ![
      "Long",
      "String",
      "Boolean",
      "Set",
      "Entity",
      "Record",
      "Extension",
      "EntityOrCommon",
    ].includes(value) &&
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
    }),
  {
    message: "Type references must be valid Cedar type names",
  }
);
