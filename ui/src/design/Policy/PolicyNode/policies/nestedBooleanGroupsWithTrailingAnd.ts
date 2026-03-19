import { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

/*
when {
  principal has email &&
  (
    (resource in Folder::"Public" && true) ||
    ((action == Action::"readFile" || resource.getTag("classification")) &&
      context == {"region": "eu-west-1"})
  ) &&
  principal.email like "*@example.com"
}
*/
const nestedBooleanGroupsWithTrailingAnd: ExpressionType = {
  "&&": {
    left: {
      has: { left: { Var: "principal" }, attr: "email" },
    },
    right: {
      "&&": {
        left: {
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
                  "||": {
                    left: {
                      "==": {
                        left: { Var: "action" },
                        right: {
                          Value: {
                            __entity: { type: "Action", id: "readFile" },
                          },
                        },
                      },
                    },
                    right: {
                      getTag: {
                        left: { Var: "resource" },
                        right: { Value: "classification" },
                      },
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
        right: {
          like: {
            left: { ".": { left: { Var: "principal" }, attr: "email" } },
            pattern: ["Wildcard", { Literal: "@example.com" }],
          },
        },
      },
    },
  },
};

export default nestedBooleanGroupsWithTrailingAnd;
