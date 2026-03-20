export type CedarAnnotations = Record<string, string>;

export type CedarPrimitiveTypeName = "Long" | "String" | "Boolean";

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

export type CedarJsonSchemaRecord = Record<
  string,
  CedarNamespaceDefinitionInput
>;
