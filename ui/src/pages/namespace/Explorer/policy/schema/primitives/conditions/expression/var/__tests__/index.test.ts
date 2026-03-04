import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Var Expression schema", () => {
  test("accepts all valid Var values", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Var: "principal" } },
          { kind: "when", body: { Var: "action" } },
          { kind: "when", body: { Var: "resource" } },
          { kind: "when", body: { Var: "context" } },
        ],
      })
    );
  });

  test("rejects invalid Var value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Var: "actor" } }],
      })
    );
  });
});
