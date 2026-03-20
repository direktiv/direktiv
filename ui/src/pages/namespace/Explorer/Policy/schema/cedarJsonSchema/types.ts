export type CedarAnnotations = Record<string, string>;

type CedarPrimitiveTypeName = "Long" | "String" | "Boolean";
type NonEmptyArray<T> = [T, ...T[]];

export type CedarActionEntityTypeName = "Action" | `${string}::Action`;

// A top-level Cedar type reference used by `shape`, `tags`, `context`, and `commonTypes`.
// Example Cedar syntax:
// - `String`
// - `Set<User>`
// - `{ owner: User, tags?: Set<String> }`
export type CedarSchemaTypeInput =
  | {
      type: CedarPrimitiveTypeName | string;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Set";
      element: CedarSchemaTypeInput;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Entity";
      name: string;
      annotations?: CedarAnnotations;
    }
  | {
      type: "Record";
      attributes: Record<string, CedarSchemaAttributeTypeInput>;
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

// Record attributes use the same type language as root types, but can also include the
// JSON-only `required` flag.
// Cedar: `name?: String`
// JSON:  `{ "type": "String", "required": false }`
export type CedarSchemaAttributeTypeInput =
  | {
      type: CedarPrimitiveTypeName | string;
      annotations?: CedarAnnotations;
      required?: boolean;
    }
  | {
      type: "Set";
      element: CedarSchemaTypeInput;
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
      attributes: Record<string, CedarSchemaAttributeTypeInput>;
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

// The full JSON schema is a map of namespace name -> namespace definition.
// The empty string represents Cedar's empty namespace.
export type CedarSchemaNamespacesInput = Record<
  string,
  CedarNamespaceDefinitionInput
>;
