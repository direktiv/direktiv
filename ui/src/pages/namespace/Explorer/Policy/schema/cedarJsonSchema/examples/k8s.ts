import type { CedarJsonSchemaInputType } from "..";

export const k8sSchema = {
  k8s: {
    entityTypes: {
      Extra: {
        shape: {
          type: "Record",
          attributes: {
            key: {
              type: "String",
              required: true,
            },
            values: {
              type: "Set",
              required: false,
              element: {
                type: "String",
              },
            },
          },
        },
      },
      Group: {
        shape: {
          type: "Record",
          attributes: {
            name: {
              type: "String",
              required: true,
            },
          },
        },
      },
      Node: {
        shape: {
          type: "Record",
          attributes: {
            extra: {
              type: "Set",
              required: false,
              element: {
                type: "ExtraAttribute",
              },
            },
            name: {
              type: "String",
              required: true,
            },
          },
        },
        memberOfTypes: ["Group"],
      },
      NonResourceURL: {
        shape: {
          type: "Record",
          attributes: {
            path: {
              type: "String",
              required: true,
            },
          },
        },
      },
      PrincipalUID: {
        shape: {
          type: "Record",
          attributes: {},
        },
      },
      Resource: {
        shape: {
          type: "Record",
          attributes: {
            apiGroup: {
              type: "String",
              required: true,
            },
            fieldSelector: {
              type: "Set",
              required: false,
              element: {
                type: "FieldRequirement",
              },
            },
            labelSelector: {
              type: "Set",
              required: false,
              element: {
                type: "LabelRequirement",
              },
            },
            name: {
              type: "String",
              required: false,
            },
            namespace: {
              type: "String",
              required: false,
            },
            resource: {
              type: "String",
              required: true,
            },
            subresource: {
              type: "String",
              required: false,
            },
          },
        },
      },
      ServiceAccount: {
        shape: {
          type: "Record",
          attributes: {
            extra: {
              type: "Set",
              required: false,
              element: {
                type: "ExtraAttribute",
              },
            },
            name: {
              type: "String",
              required: true,
            },
            namespace: {
              type: "String",
              required: true,
            },
          },
        },
        memberOfTypes: ["Group"],
      },
      User: {
        shape: {
          type: "Record",
          attributes: {
            extra: {
              type: "Set",
              required: false,
              element: {
                type: "ExtraAttribute",
              },
            },
            name: {
              type: "String",
              required: true,
            },
          },
        },
        memberOfTypes: ["Group"],
      },
    },
    actions: {
      approve: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      attest: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      bind: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      create: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      delete: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL", "Resource"],
        },
      },
      deletecollection: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      escalate: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      get: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL", "Resource"],
        },
        memberOf: [
          {
            id: "readOnly",
          },
        ],
      },
      head: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL"],
        },
      },
      impersonate: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: [
            "Extra",
            "Group",
            "Node",
            "PrincipalUID",
            "ServiceAccount",
            "User",
          ],
        },
      },
      list: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
        memberOf: [
          {
            id: "readOnly",
          },
        ],
      },
      options: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL"],
        },
      },
      patch: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL", "Resource"],
        },
      },
      post: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL"],
        },
      },
      put: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["NonResourceURL"],
        },
      },
      readOnly: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      sign: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      update: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      use: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
      watch: {
        appliesTo: {
          principalTypes: ["Group", "Node", "ServiceAccount", "User"],
          resourceTypes: ["Resource"],
        },
      },
    },
    commonTypes: {
      ExtraAttribute: {
        type: "Record",
        attributes: {
          key: {
            type: "String",
            required: true,
          },
          values: {
            type: "Set",
            required: false,
            element: {
              type: "String",
            },
          },
        },
      },
      FieldRequirement: {
        type: "Record",
        attributes: {
          field: {
            type: "String",
            required: true,
          },
          operator: {
            type: "String",
            required: true,
          },
          value: {
            type: "String",
            required: true,
          },
        },
      },
      LabelRequirement: {
        type: "Record",
        attributes: {
          key: {
            type: "String",
            required: true,
          },
          operator: {
            type: "String",
            required: true,
          },
          values: {
            type: "Set",
            required: true,
            element: {
              type: "String",
            },
          },
        },
      },
    },
  },
} satisfies CedarJsonSchemaInputType;
