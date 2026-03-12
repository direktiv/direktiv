import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, expect, test } from "vitest";
import { CedarPolicySchema } from "../../..";

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
    createBasePolicy({
      // @ts-expect-error - conditions require a body expression
      conditions: [{ kind: "when" }],
    });

    expect(
      CedarPolicySchema.safeParse({
        ...createBasePolicy(),
        conditions: [{ kind: "when" }],
      }).success
    ).toBe(false);
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
