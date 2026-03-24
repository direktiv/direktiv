export type CedarAnnotations = Record<string, string>;

export const cedarPrimitiveTypeNames = ["Long", "String", "Boolean"] as const;
type CedarPrimitiveTypeName = (typeof cedarPrimitiveTypeNames)[number];
type NonEmptyArray<T> = [T, ...T[]];

export type CedarActionEntityTypeName = "Action" | `${string}::Action`;

// Base metadata shared by every node in the schema tree
type CedarSchemaTypeMetadata = {
  annotations?: CedarAnnotations;
};

// Record attributes use the same type grammar as top-level schema types, but add
// the JSON schema `required` flag because required only applies to attributes.
type CedarSchemaAttributeMetadata = CedarSchemaTypeMetadata & {
  required?: boolean;
};

// These variants describe the recursive Cedar type grammar used by `shape`,
// `tags`, `context`, and namespace `commonTypes`.
type CedarPrimitiveOrCommonType = CedarSchemaTypeMetadata & {
  type: CedarPrimitiveTypeName | string;
};

type CedarSetType = CedarSchemaTypeMetadata & {
  type: "Set";
  element: CedarSchemaType;
};

type CedarEntityReferenceType = CedarSchemaTypeMetadata & {
  type: "Entity";
  name: string;
};

type CedarRecordType = CedarSchemaTypeMetadata & {
  type: "Record";
  attributes: Record<string, CedarSchemaAttributeType>;
};

type CedarExtensionType = CedarSchemaTypeMetadata & {
  type: "Extension";
  name: string;
};

type CedarEntityOrCommonType = CedarSchemaTypeMetadata & {
  type: "EntityOrCommon";
  name: string;
};

// Attribute variants intentionally mirror the schema variants above.
// The only difference is that each node carries attribute metadata so nested
// record fields can express `required`.
type CedarAttributePrimitiveOrCommonType = CedarSchemaAttributeMetadata & {
  type: CedarPrimitiveTypeName | string;
};

type CedarAttributeSetType = CedarSchemaAttributeMetadata & {
  type: "Set";
  element: CedarSchemaType;
};

type CedarAttributeEntityReferenceType = CedarSchemaAttributeMetadata & {
  type: "Entity";
  name: string;
};

type CedarAttributeRecordType = CedarSchemaAttributeMetadata & {
  type: "Record";
  attributes: Record<string, CedarSchemaAttributeType>;
};

type CedarAttributeExtensionType = CedarSchemaAttributeMetadata & {
  type: "Extension";
  name: string;
};

type CedarAttributeEntityOrCommonType = CedarSchemaAttributeMetadata & {
  type: "EntityOrCommon";
  name: string;
};

// `CedarSchemaType` is the core recursive union for any standalone Cedar type.
export type CedarSchemaType =
  | CedarPrimitiveOrCommonType
  | CedarSetType
  | CedarEntityReferenceType
  | CedarRecordType
  | CedarExtensionType
  | CedarEntityOrCommonType;

// Attribute counterpart to `CedarSchemaType`.
// It exists so record attributes can reuse the same nested shapes while adding
// the optional `required` flag at every attribute node.
export type CedarSchemaAttributeType =
  | CedarAttributePrimitiveOrCommonType
  | CedarAttributeSetType
  | CedarAttributeEntityReferenceType
  | CedarAttributeRecordType
  | CedarAttributeExtensionType
  | CedarAttributeEntityOrCommonType;

// Namespace `entityTypes` map to one of these two JSON shapes:
// - structural:
//   {
//     "memberOfTypes": ["Group"],
//     "shape": {
//       "type": "Record",
//       "attributes": {
//         "name": {
//           "type": "String",
//           "required": true
//         }
//       }
//     }
//   }
// - enum:
//   {
//     "enum": ["admins", "reviewers"]
//   }
// Structural entities plug back into `CedarSchemaType` for `shape` and `tags`.
export type CedarEntityDefinition =
  | {
      memberOfTypes?: string[];
      shape?: CedarSchemaType;
      tags?: CedarSchemaType;
      annotations?: CedarAnnotations;
    }
  | {
      enum: NonEmptyArray<string>;
      annotations?: CedarAnnotations;
    };

export type CedarActionGroupReference = {
  id: string;
  type?: CedarActionEntityTypeName;
};

export type CedarActionDeclaration = {
  memberOf?: CedarActionGroupReference[];
  appliesTo: {
    principalTypes: string[];
    resourceTypes: string[];
    context?: CedarSchemaType;
  };
  annotations?: CedarAnnotations;
};

export type CedarNamespaceDefinition = {
  entityTypes: Record<string, CedarEntityDefinition>;
  actions: Record<string, CedarActionDeclaration>;
  commonTypes?: Record<string, CedarSchemaType>;
  annotations?: CedarAnnotations;
};

export type CedarSchemaNamespaces = Record<string, CedarNamespaceDefinition>;
