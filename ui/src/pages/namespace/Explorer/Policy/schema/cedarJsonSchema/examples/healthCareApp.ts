import type { CedarSchemaNamespaces } from "../types";

export const healthCareAppSchema = {
  HealthCareApp: {
    entityTypes: {
      User: {
        shape: {
          type: "Record",
          attributes: {},
        },
        memberOfTypes: ["Role"],
      },
      Role: {
        shape: {
          type: "Record",
          attributes: {},
        },
        memberOfTypes: [],
      },
      Info: {
        shape: {
          type: "Record",
          attributes: {
            provider: {
              type: "Entity",
              name: "User",
            },
            patient: {
              type: "Entity",
              name: "User",
            },
          },
        },
        memberOfTypes: ["InfoType"],
      },
      InfoType: {
        shape: {
          type: "Record",
          attributes: {},
        },
        memberOfTypes: [],
      },
    },
    actions: {
      createAppointment: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["Info"],
          context: {
            type: "Record",
            attributes: {
              referrer: {
                type: "Entity",
                name: "User",
              },
            },
          },
        },
      },
      updateAppointment: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["Info"],
          context: {
            type: "Record",
            attributes: {},
          },
        },
      },
      deleteAppointment: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["Info"],
          context: {
            type: "Record",
            attributes: {},
          },
        },
      },
      listAppointments: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["Info"],
          context: {
            type: "Record",
            attributes: {},
          },
        },
      },
    },
  },
} satisfies CedarSchemaNamespaces;
