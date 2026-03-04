import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Like Expression schema", () => {
  test("accepts like expression", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              like: {
                left: { ".": { left: { Var: "resource" }, attr: "email" } },
                pattern: ["Wildcard", { Literal: "@amazon.com" }],
              },
            },
          },
        ],
      })
    );
  });

  test("rejects like expression with invalid pattern element", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { like: { left: { Var: "resource" }, pattern: [123] } },
          },
        ],
      })
    );
  });

  test("rejects non-strict literal pattern shape", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              like: {
                left: { Var: "resource" },
                pattern: [{ Literal: "mail", extra: true }],
              },
            },
          },
        ],
      })
    );
  });
});
