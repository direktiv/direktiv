import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../testUtils";
import { describe, test } from "vitest";

describe("Cedar effect schema", () => {
  test("accepts permit effect", () => {
    expectValidPolicy(
      createBasePolicy({
        effect: "permit",
      })
    );
  });

  test("accepts forbid effect", () => {
    expectValidPolicy(
      createBasePolicy({
        effect: "forbid",
      })
    );
  });

  test("rejects unknown effect", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - only permit/forbid are allowed
        effect: "allow",
      })
    );
  });
});
