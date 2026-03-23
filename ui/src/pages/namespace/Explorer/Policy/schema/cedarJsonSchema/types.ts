export type CedarAnnotations = Record<string, string>;

export const cedarPrimitiveTypeNames = ["Long", "String", "Boolean"] as const;
type CedarPrimitiveTypeName = (typeof cedarPrimitiveTypeNames)[number];
type NonEmptyArray<T> = [T, ...T[]];

export type CedarActionEntityTypeName = "Action" | `${string}::Action`;

type CedarSchemaTypeMetadata = {
  annotations?: CedarAnnotations;
};

type CedarSchemaAttributeMetadata = CedarSchemaTypeMetadata & {
  required?: boolean;
};

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

// `shape`, `tags`, `context`, and common types all use the same recursive Cedar
// type grammar. Record attributes reuse that grammar and add the JSON-only
// `required` flag.
export type CedarSchemaType =
  | CedarPrimitiveOrCommonType
  | CedarSetType
  | CedarEntityReferenceType
  | CedarRecordType
  | CedarExtensionType
  | CedarEntityOrCommonType;

export type CedarSchemaAttributeType =
  | CedarAttributePrimitiveOrCommonType
  | CedarAttributeSetType
  | CedarAttributeEntityReferenceType
  | CedarAttributeRecordType
  | CedarAttributeExtensionType
  | CedarAttributeEntityOrCommonType;

// Entity types come in two shapes:
// - structural entities: `entity User in Group = { name: String };`
// - enum entities: `entity Group enum ["admins", "reviewers"];`
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

// Cedar action declarations are centered around `appliesTo`.
// Cedar:
// `action View appliesTo { principal: User, resource: Doc, context: { ip: ipaddr } };`
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
