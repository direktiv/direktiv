import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

/*
  principal has email && principal.email like "*@example.com"
*/
const simpleWhen: ExpressionType = {
  "&&": {
    left: { has: { left: { Var: "principal" }, attr: "email" } },
    right: {
      like: {
        left: { ".": { left: { Var: "principal" }, attr: "email" } },
        pattern: ["Wildcard", { Literal: "@example.com" }],
      },
    },
  },
};

export default simpleWhen;
