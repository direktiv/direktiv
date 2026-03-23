import { CedarSchemaNamespacesInput } from "../types";

export const awsSchema = {
  "AWS::IdentityStore": {
    entityTypes: {
      Group: {},
      User: {
        memberOfTypes: ["Group"],
        shape: {
          type: "Record",
          attributes: {
            costCenter: {
              type: "EntityOrCommon",
              name: "String",
              required: false,
            },
            division: {
              type: "EntityOrCommon",
              name: "String",
              required: false,
            },
            employeeNumber: {
              type: "EntityOrCommon",
              name: "String",
              required: false,
            },
            organization: {
              type: "EntityOrCommon",
              name: "String",
              required: false,
            },
          },
        },
      },
    },
    actions: {},
  },
  "AWS::SSM": {
    entityTypes: {
      ManagedInstance: {
        tags: {
          type: "EntityOrCommon",
          name: "String",
        },
      },
    },
    actions: {
      getTokenForInstanceAccess: {
        appliesTo: {
          resourceTypes: ["AWS::EC2::Instance", "AWS::SSM::ManagedInstance"],
          principalTypes: ["AWS::IdentityStore::User"],
          context: {
            type: "Record",
            attributes: {
              iam: {
                type: "EntityOrCommon",
                name: "AWS::IAM::AuthorizationContext",
              },
            },
          },
        },
      },
    },
  },
  "AWS::EC2": {
    entityTypes: {
      Instance: {
        tags: {
          type: "EntityOrCommon",
          name: "String",
        },
      },
    },
    actions: {},
  },
  "AWS::IAM": {
    commonTypes: {
      AuthorizationContext: {
        type: "Record",
        attributes: {
          principalTags: {
            type: "EntityOrCommon",
            name: "PrincipalTags",
          },
        },
      },
    },
    entityTypes: {
      PrincipalTags: {
        tags: {
          type: "EntityOrCommon",
          name: "String",
        },
      },
      Role: {
        tags: {
          type: "EntityOrCommon",
          name: "String",
        },
      },
    },
    actions: {},
  },
} satisfies CedarSchemaNamespacesInput;
