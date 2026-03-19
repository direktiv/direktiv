import {
  createBasePolicy,
  expectValidPolicy,
} from "../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar Expression schema", () => {
  test("accepts nested Expression variants", () => {
    // Cedar: when { if !context then resource.email like "*@amazon.com" else false };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "if-then-else": {
                if: { "!": { arg: { Var: "context" } } },
                then: {
                  like: {
                    left: { ".": { left: { Var: "resource" }, attr: "email" } },
                    pattern: ["Wildcard", { Literal: "@amazon.com" }],
                  },
                },
                else: { Value: false },
              },
            },
          },
        ],
      })
    );
  });
});
