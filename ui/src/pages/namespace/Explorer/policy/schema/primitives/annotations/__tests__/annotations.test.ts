import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../testUtils";
import { describe, test } from "vitest";

describe("Cedar annotations schema", () => {
  test("accepts annotations with string and null", () => {
    const input = createBasePolicy({
      annotations: {
        shadow_mode: null,
        reason: "temporary block",
      },
    });

    expectValidPolicy(input);
  });

  test("rejects invalid annotation value type", () => {
    expectInvalidPolicy(
      createBasePolicy({
        annotations: {
          // @ts-expect-error - annotation values must be string or null
          priority: 10,
        },
      })
    );
  });
});
