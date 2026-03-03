import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar resource schema", () => {
  test("accepts resource All operator", () => {
    expectValidPolicy(createBasePolicy({ resource: { op: "All" } }));
  });

  test("accepts resource == entity", () => {
    const input = createBasePolicy({
      resource: { op: "==", entity: { type: "Folder", id: "Public" } },
    });
    expectValidPolicy(input);
  });

  test("accepts resource == slot", () => {
    const input = createBasePolicy({
      resource: { op: "==", slot: "?resource" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource in entity", () => {
    const input = createBasePolicy({
      resource: { op: "in", entity: { type: "Folder", id: "Public" } },
    });
    expectValidPolicy(input);
  });

  test("accepts resource in slot", () => {
    const input = createBasePolicy({
      resource: { op: "in", slot: "?resource" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource is entity type", () => {
    const input = createBasePolicy({
      resource: { op: "is", entity_type: "Folder" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource is in entity", () => {
    const input = createBasePolicy({
      resource: {
        op: "is",
        entity_type: "Folder",
        in: { entity: { type: "Folder", id: "Public" } },
      },
    });
    expectValidPolicy(input);
  });

  test("accepts resource is in slot", () => {
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

  test("rejects invalid resource in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - resource slot only allows ?resource
        resource: { op: "in", slot: "?principal" },
      })
    );
  });

  test("rejects resource is variant with invalid in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - resource is/in slot only allows ?resource
        resource: {
          op: "is",
          entity_type: "Folder",
          in: { slot: "?principal" },
        },
      })
    );
  });

  test("rejects resource == variant with missing entity or slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - resource == requires entity or slot
        resource: { op: "==" },
      })
    );
  });
});
