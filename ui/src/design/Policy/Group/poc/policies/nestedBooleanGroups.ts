import type { ConditionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";

const nestedBooleanGroups: ConditionType[] = [
  {
    kind: "when",
    body: {
      "&&": {
        left: { has: { left: { Var: "principal" }, attr: "email" } },
        right: {
          "||": {
            left: {
              "&&": {
                left: {
                  in: {
                    left: { Var: "resource" },
                    right: {
                      Value: { __entity: { type: "Folder", id: "Public" } },
                    },
                  },
                },
                right: { Value: true },
              },
            },
            right: {
              "&&": {
                left: {
                  "==": {
                    left: { Var: "action" },
                    right: {
                      Value: { __entity: { type: "Action", id: "readFile" } },
                    },
                  },
                },
                right: {
                  "==": {
                    left: { Var: "context" },
                    right: { Value: { region: "eu-west-1" } },
                  },
                },
              },
            },
          },
        },
      },
    },
  },
];

export default nestedBooleanGroups;
