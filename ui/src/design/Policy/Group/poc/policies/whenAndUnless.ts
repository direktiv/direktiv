import type { ConditionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";

import simpleWhen from "./simpleWhen";

const whenAndUnless: ConditionType[] = [
  ...simpleWhen,
  {
    kind: "unless",
    body: {
      "||": {
        left: {
          getTag: {
            left: { Var: "resource" },
            right: { Value: "classification" },
          },
        },
        right: {
          "&&": {
            left: { Value: false },
            right: {
              is: {
                left: { Var: "principal" },
                entity_type: "User",
              },
            },
          },
        },
      },
    },
  },
];

export default whenAndUnless;
