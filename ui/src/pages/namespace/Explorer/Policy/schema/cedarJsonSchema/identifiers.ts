import { z } from "zod";

const cedarReservedTypeNames = new Set([
  "Bool",
  "Boolean",
  "Entity",
  "Extension",
  "Long",
  "Record",
  "Set",
  "String",
]);

const cedarIdentifierSegment = /^[_a-zA-Z][_a-zA-Z0-9]*$/;

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

  if (segments.some((segment) => !cedarIdentifierSegment.test(segment))) {
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

  return allowReservedFinalSegment || !cedarReservedTypeNames.has(finalSegment);
};

export const NamespaceNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value, { allowEmpty: true }), {
    message: "Namespace names must be empty or valid Cedar identifier paths",
  });

export const EntityTypeNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value), {
    message: "Entity type names must be valid Cedar identifier paths",
  });

export const CommonTypeReferenceSchema = z.string().refine(
  (value) =>
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
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

export const ActionEntityTypeSchema = z
  .string()
  .refine(
    (value) => isIdentifierPath(value, { allowReservedFinalSegment: true }),
    {
      message: "Action entity types must be valid Cedar identifier paths",
    }
  )
  .refine((value) => value.split("::").at(-1) === "Action", {
    message: "Action entity types must end with 'Action'",
  });

export const PrimitiveTypeNameSchema = z.enum(["Long", "String", "Boolean"]);

export const TypeReferenceSchema = z.string().refine(
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
