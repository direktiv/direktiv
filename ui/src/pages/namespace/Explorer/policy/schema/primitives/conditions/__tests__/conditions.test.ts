import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar conditions schema", () => {
  test("accepts multiple conditions", () => {
    // Cedar: when { true } unless { false };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Value: true } },
          { kind: "unless", body: { Value: false } },
        ],
      })
    );
  });

  test("accepts when condition", () => {
    // Cedar: when { true };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Value: true } }],
      })
    );
  });

  test("accepts unless condition", () => {
    // Cedar: unless { true };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "unless", body: { Value: true } }],
      })
    );
  });

  test("rejects invalid condition kind", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            // @ts-expect-error - condition kind only allows when/unless
            kind: "if",
            body: { Value: true },
          },
        ],
      })
    );
  });

  test("rejects condition without body", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when" }],
      })
    );
  });

  test("rejects condition with additional keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { Value: true },
            // @ts-expect-error - condition is strict and disallows extra keys
            extra: true,
          },
        ],
      })
    );
  });
});
