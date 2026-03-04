import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar effect schema", () => {
  test("accepts permit effect", () => {
    // Cedar: permit(principal, action, resource);
    expectValidPolicy(
      createBasePolicy({
        effect: "permit",
      })
    );
  });

  test("accepts forbid effect", () => {
    // Cedar: forbid(principal, action, resource);
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
