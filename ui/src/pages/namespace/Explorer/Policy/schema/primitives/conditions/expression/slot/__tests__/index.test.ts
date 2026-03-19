import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Slot Expression schema", () => {
  test("accepts all valid Slot values", () => {
    /*
      Cedar templates:
      when { principal == ?principal };
      when { resource == ?resource };
    */
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { Var: "principal" },
                right: { Slot: "?principal" },
              },
            },
          },
          {
            kind: "when",
            body: {
              "==": {
                left: { Var: "resource" },
                right: { Slot: "?resource" },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects invalid Slot value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - slot only allows ?principal and ?resource
        conditions: [{ kind: "when", body: { Slot: "?action" } }],
      })
    );
  });
});
