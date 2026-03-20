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

const isIdentifierPath = (
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

const recordWithValidatedKeys = <Schema extends z.ZodTypeAny>(
  valueSchema: Schema,
  isValidKey: (key: string) => boolean,
  message: string
) =>
  z.record(valueSchema).superRefine((value, ctx) => {
    Object.keys(value).forEach((key) => {
      if (!isValidKey(key)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message,
          path: [key],
        });
      }
    });
  });

const AnnotationsSchema = z.record(z.string());

const NamespaceNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value, { allowEmpty: true }), {
    message: "Namespace names must be empty or valid Cedar identifier paths",
  });

const EntityTypeNameSchema = z
  .string()
  .refine((value) => isIdentifierPath(value), {
    message: "Entity type names must be valid Cedar identifier paths",
  });

const CommonTypeReferenceSchema = z.string().refine(
  (value) =>
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
    }),
  {
    message: "Common type references must be valid Cedar identifier paths",
  }
);

const ExtensionTypeNameSchema = z.string().refine(
  (value) =>
    isIdentifierPath(value, {
      allowReservedCedarNamespace: value.startsWith("__cedar::"),
      allowReservedFinalSegment: true,
    }),
  {
    message: "Extension type names must be valid Cedar identifier paths",
  }
);

const ActionEntityTypeSchema = z
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

const PrimitiveTypeNameSchema = z.enum(["Long", "String", "Boolean"]);

type CedarAnnotations = Record<string, string>;

type CedarPrimitiveTypeName = z.infer<typeof PrimitiveTypeNameSchema>;

type CedarRootTypeInput =
  | {
      type: CedarPrimitiveTypeName | string;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Set";
      element: CedarRootTypeInput;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Entity";
      name: string;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Record";
      attributes: Record<string, CedarRecordAttributeInput>;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Extension";
      name: string;
      annotations?: CedarAnnotations;
    }
  | {
      type: "EntityOrCommon";
      name: string;
      annotations?: CedarAnnotations;
    };

type CedarRecordAttributeInput =
  | {
      type: CedarPrimitiveTypeName | string;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "Set";
      element: CedarRootTypeInput;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "Entity";
      name: string;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "Record";
      attributes: Record<string, CedarRecordAttributeInput>;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "Extension";
      name: string;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "EntityOrCommon";
      name: string;
      annotations?: CedarAnnotations;
      required?: boolean;
    };

type CedarEntityTypeDefinitionInput =
  | {
      memberOfTypes?: string[];
      shape?: CedarRootTypeInput;
      tags?: CedarRootTypeInput;
      annotations?: CedarAnnotations;
    }
  | {
      enum: string[];
      annotations?: CedarAnnotations;
    };

type CedarActionGroupMembershipInput = {
  id: string;
  type?: string;
};

type CedarActionDefinitionInput = {
  memberOf?: CedarActionGroupMembershipInput[];
  appliesTo: {
    principalTypes: string[];
    resourceTypes: string[];
    context?: CedarRootTypeInput;
  };
  annotations?: CedarAnnotations;
};

type CedarNamespaceDefinitionInput = {
  entityTypes: Record<string, CedarEntityTypeDefinitionInput>;
  actions: Record<string, CedarActionDefinitionInput>;
  commonTypes?: Record<string, CedarRootTypeInput>;
  annotations?: CedarAnnotations;
};

const TypeReferenceSchema = z.string().refine(
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

RootTypeSchema = buildTypeSchema({ allowRequired: false });
RecordAttributeSchema = buildTypeSchema({ allowRequired: true });

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

export const CedarJsonSchema: z.ZodType<
  Record<string, CedarNamespaceDefinitionInput>
> = recordWithValidatedKeys(
  NamespaceDefinitionSchema,
  (key) => NamespaceNameSchema.safeParse(key).success,
  "Namespace names must be empty or valid Cedar identifier paths"
).refine((value) => Object.keys(value).length > 0, {
  message: "A Cedar schema must declare at least one namespace",
});

export type CedarJsonSchemaType = z.infer<typeof CedarJsonSchema>;
export type CedarJsonSchemaInputType = z.input<typeof CedarJsonSchema>;
