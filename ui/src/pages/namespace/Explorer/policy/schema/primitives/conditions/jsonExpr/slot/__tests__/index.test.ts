import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Slot JsonExpr schema", () => {
  test("accepts all valid Slot values", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Slot: "?principal" } },
          { kind: "when", body: { Slot: "?resource" } },
        ],
      })
    );
  });

  test("rejects invalid Slot value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Slot: "?action" } }],
      })
    );
  });
});
