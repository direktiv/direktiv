export type CedarAnnotations = Record<string, string>;

type CedarPrimitiveTypeName = "Long" | "String" | "Boolean";

// A top-level Cedar type reference used by `shape`, `tags`, `context`, and `commonTypes`.
// Example Cedar syntax:
// - `String`
// - `Set<User>`
// - `{ owner: User, tags?: Set<String> }`
export type CedarRootTypeInput =
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

// Record attributes use the same type language as root types, but can also include the
// JSON-only `required` flag.
// Cedar: `name?: String`
// JSON:  `{ "type": "String", "required": false }`
export type CedarRecordAttributeInput =
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

// Entity types come in two shapes:
// - structural entities: `entity User in Group = { name: String };`
// - enum entities: `entity Group enum ["admins", "reviewers"];`
export type CedarEntityTypeDefinitionInput =
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

export type CedarActionGroupMembershipInput = {
  id: string;
  type?: string;
};

// Cedar action declarations are centered around `appliesTo`.
// Cedar:
// `action View appliesTo { principal: User, resource: Doc, context: { ip: ipaddr } };`
export type CedarActionDefinitionInput = {
  memberOf?: CedarActionGroupMembershipInput[];
  appliesTo: {
    principalTypes: string[];
    resourceTypes: string[];
    context?: CedarRootTypeInput;
  };
  annotations?: CedarAnnotations;
};

export type CedarNamespaceDefinitionInput = {
  entityTypes: Record<string, CedarEntityTypeDefinitionInput>;
  actions: Record<string, CedarActionDefinitionInput>;
  commonTypes?: Record<string, CedarRootTypeInput>;
  annotations?: CedarAnnotations;
};

// The full JSON schema is a map of namespace name -> namespace definition.
// The empty string represents Cedar's empty namespace.
export type CedarJsonSchemaRecord = Record<
  string,
  CedarNamespaceDefinitionInput
>;
