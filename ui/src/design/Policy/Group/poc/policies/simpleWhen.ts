import type { ConditionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";

const simpleWhen: ConditionType[] = [
  {
    kind: "when",
    body: {
      "&&": {
        left: { has: { left: { Var: "principal" }, attr: "email" } },
        right: {
          like: {
            left: { ".": { left: { Var: "principal" }, attr: "email" } },
            pattern: ["Wildcard", { Literal: "@example.com" }],
          },
        },
      },
    },
  },
];

export default simpleWhen;
