import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Attribute JsonExpr schema", () => {
  test("accepts dot accessor expression", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { ".": { left: { Var: "context" }, attr: "tls_version" } },
          },
        ],
      })
    );
  });

  test("accepts has accessor expression", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { has: { left: { Var: "principal" }, attr: "email" } },
          },
        ],
      })
    );
  });

  test("rejects attribute expression without attr", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { ".": { left: { Var: "context" } } } },
        ],
      })
    );
  });

  test("rejects non-string attr value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { has: { left: { Var: "principal" }, attr: 1 } },
          },
        ],
      })
    );
  });
});
