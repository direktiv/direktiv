import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar conditions schema", () => {
  test("accepts when condition", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Value: true } }],
      })
    );
  });

  test("accepts unless condition", () => {
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
});
