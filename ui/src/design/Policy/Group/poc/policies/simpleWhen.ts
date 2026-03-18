import type { ExpressionInput } from "../utils";

/*
  principal has email && principal.email like "*@example.com"
*/
const simpleWhen: ExpressionInput = {
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
