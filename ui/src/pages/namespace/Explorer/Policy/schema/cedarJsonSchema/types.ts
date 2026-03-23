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

type CedarPrimitiveOrCommonTypeInput = CedarSchemaTypeMetadata & {
  type: CedarPrimitiveTypeName | string;
};

type CedarSetTypeInput = CedarSchemaTypeMetadata & {
  type: "Set";
  element: CedarSchemaTypeInput;
};

type CedarEntityReferenceTypeInput = CedarSchemaTypeMetadata & {
  type: "Entity";
  name: string;
};

type CedarRecordTypeInput = CedarSchemaTypeMetadata & {
  type: "Record";
  attributes: Record<string, CedarSchemaAttributeTypeInput>;
};

type CedarExtensionTypeInput = CedarSchemaTypeMetadata & {
  type: "Extension";
  name: string;
};

type CedarEntityOrCommonTypeInput = CedarSchemaTypeMetadata & {
  type: "EntityOrCommon";
  name: string;
};

type CedarAttributePrimitiveOrCommonTypeInput = CedarSchemaAttributeMetadata & {
  type: CedarPrimitiveTypeName | string;
};

type CedarAttributeSetTypeInput = CedarSchemaAttributeMetadata & {
  type: "Set";
  element: CedarSchemaTypeInput;
};

type CedarAttributeEntityReferenceTypeInput = CedarSchemaAttributeMetadata & {
  type: "Entity";
  name: string;
};

type CedarAttributeRecordTypeInput = CedarSchemaAttributeMetadata & {
  type: "Record";
  attributes: Record<string, CedarSchemaAttributeTypeInput>;
};

type CedarAttributeExtensionTypeInput = CedarSchemaAttributeMetadata & {
  type: "Extension";
  name: string;
};

type CedarAttributeEntityOrCommonTypeInput = CedarSchemaAttributeMetadata & {
  type: "EntityOrCommon";
  name: string;
};

// `shape`, `tags`, `context`, and common types all use the same recursive Cedar
// type grammar. Record attributes reuse that grammar and add the JSON-only
// `required` flag.
export type CedarSchemaTypeInput =
  | CedarPrimitiveOrCommonTypeInput
  | CedarSetTypeInput
  | CedarEntityReferenceTypeInput
  | CedarRecordTypeInput
  | CedarExtensionTypeInput
  | CedarEntityOrCommonTypeInput;

export type CedarSchemaAttributeTypeInput =
  | CedarAttributePrimitiveOrCommonTypeInput
  | CedarAttributeSetTypeInput
  | CedarAttributeEntityReferenceTypeInput
  | CedarAttributeRecordTypeInput
  | CedarAttributeExtensionTypeInput
  | CedarAttributeEntityOrCommonTypeInput;

// Entity types come in two shapes:
// - structural entities: `entity User in Group = { name: String };`
// - enum entities: `entity Group enum ["admins", "reviewers"];`
export type CedarEntityDefinitionInput =
  | {
      memberOfTypes?: string[];
      shape?: CedarSchemaTypeInput;
      tags?: CedarSchemaTypeInput;
      annotations?: CedarAnnotations;
    }
  | {
      enum: NonEmptyArray<string>;
      annotations?: CedarAnnotations;
    };

export type CedarActionGroupReferenceInput = {
  id: string;
  type?: CedarActionEntityTypeName;
};

// Cedar action declarations are centered around `appliesTo`.
// Cedar:
// `action View appliesTo { principal: User, resource: Doc, context: { ip: ipaddr } };`
export type CedarActionDeclarationInput = {
  memberOf?: CedarActionGroupReferenceInput[];
  appliesTo: {
    principalTypes: string[];
    resourceTypes: string[];
    context?: CedarSchemaTypeInput;
  };
  annotations?: CedarAnnotations;
};

export type CedarNamespaceDefinitionInput = {
  entityTypes: Record<string, CedarEntityDefinitionInput>;
  actions: Record<string, CedarActionDeclarationInput>;
  commonTypes?: Record<string, CedarSchemaTypeInput>;
  annotations?: CedarAnnotations;
};

export type CedarSchemaNamespacesInput = Record<
  string,
  CedarNamespaceDefinitionInput
>;
