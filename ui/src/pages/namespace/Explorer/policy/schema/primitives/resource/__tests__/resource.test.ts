import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../testUtils";
import { describe, test } from "vitest";

describe("Cedar resource schema", () => {
  test("accepts resource is with in slot", () => {
    const input = createBasePolicy({
      resource: {
        op: "is",
        entity_type: "Folder",
        in: { slot: "?resource" },
      },
    });

    expectValidPolicy(input);
  });

  test("rejects invalid resource slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - resource slot only allows ?resource
        resource: { op: "==", slot: "?principal" },
      })
    );
  });

  test("rejects resource is variant with invalid in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        resource: {
          op: "is",
          entity_type: "Folder",
          // @ts-expect-error - resource is/in slot only allows ?resource
          in: { slot: "?principal" },
        },
      })
    );
  });
});
