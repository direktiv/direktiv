import type { CedarSchemaNamespaces } from "../types";

export const minimalSchema = {
  Demo: {
    entityTypes: {
      Document: {},
      User: {
        shape: {
          type: "Record",
          attributes: {
            name: {
              type: "String",
            },
          },
        },
      },
    },
    actions: {
      viewDocument: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["Document"],
          context: {
            type: "Record",
            attributes: {
              authenticated: {
                type: "Boolean",
                required: true,
              },
            },
          },
        },
      },
    },
  },
} satisfies CedarSchemaNamespaces;
