import type { CedarJsonSchemaInputType } from "..";

export const photoAppSchema = {
  PhotoApp: {
    commonTypes: {
      PersonType: {
        type: "Record",
        attributes: {
          age: {
            type: "Long",
          },
          name: {
            type: "String",
          },
        },
      },
      ContextType: {
        type: "Record",
        attributes: {
          ip: {
            type: "Extension",
            name: "ipaddr",
            required: false,
          },
          authenticated: {
            type: "Boolean",
            required: true,
          },
        },
      },
    },
    entityTypes: {
      User: {
        shape: {
          type: "Record",
          attributes: {
            userId: {
              type: "String",
            },
            personInformation: {
              type: "PersonType",
            },
          },
        },
        memberOfTypes: ["UserGroup"],
      },
      UserGroup: {
        shape: {
          type: "Record",
          attributes: {},
        },
      },
      Photo: {
        shape: {
          type: "Record",
          attributes: {
            account: {
              type: "Entity",
              name: "Account",
              required: true,
            },
            private: {
              type: "Boolean",
              required: true,
            },
          },
        },
        memberOfTypes: ["Album", "Account"],
      },
      Album: {
        shape: {
          type: "Record",
          attributes: {},
        },
      },
      Account: {
        shape: {
          type: "Record",
          attributes: {},
        },
      },
    },
    actions: {
      viewPhoto: {
        appliesTo: {
          principalTypes: ["User", "UserGroup"],
          resourceTypes: ["Photo"],
          context: {
            type: "ContextType",
          },
        },
      },
      createPhoto: {
        appliesTo: {
          principalTypes: ["User", "UserGroup"],
          resourceTypes: ["Photo"],
          context: {
            type: "ContextType",
          },
        },
      },
      listPhotos: {
        appliesTo: {
          principalTypes: ["User", "UserGroup"],
          resourceTypes: ["Photo"],
          context: {
            type: "ContextType",
          },
        },
      },
    },
  },
} satisfies CedarJsonSchemaInputType;
