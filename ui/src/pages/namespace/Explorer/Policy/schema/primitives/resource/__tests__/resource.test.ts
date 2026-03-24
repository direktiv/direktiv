import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar resource schema", () => {
  test("accepts resource All operator", () => {
    // Cedar: permit(principal, action, resource);
    expectValidPolicy(createBasePolicy({ resource: { op: "All" } }));
  });

  test("accepts resource == entity", () => {
    // Cedar: permit(principal, action, resource == Folder::"Public");
    const input = createBasePolicy({
      resource: { op: "==", entity: { type: "Folder", id: "Public" } },
    });
    expectValidPolicy(input);
  });

  test("accepts resource == slot", () => {
    // Cedar template: permit(principal, action, resource == ?resource);
    const input = createBasePolicy({
      resource: { op: "==", slot: "?resource" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource in entity", () => {
    // Cedar: permit(principal, action, resource in Folder::"Public");
    const input = createBasePolicy({
      resource: { op: "in", entity: { type: "Folder", id: "Public" } },
    });
    expectValidPolicy(input);
  });

  test("accepts resource in slot", () => {
    // Cedar template: permit(principal, action, resource in ?resource);
    const input = createBasePolicy({
      resource: { op: "in", slot: "?resource" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource is entity type", () => {
    // Cedar: permit(principal, action, resource is Folder);
    const input = createBasePolicy({
      resource: { op: "is", entity_type: "Folder" },
    });
    expectValidPolicy(input);
  });

  test("accepts resource is in entity", () => {
    // Cedar: permit(principal, action, resource is Folder in Folder::"Public");
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
    // Cedar template: permit(principal, action, resource is Folder in ?resource);
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
        // Cedar (invalid for resource constraint): permit(principal, action, resource == ?principal);
        // @ts-expect-error - resource slot only allows ?resource
        resource: { op: "==", slot: "?principal" },
      })
    );
  });

  test("rejects invalid resource in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid for resource constraint): permit(principal, action, resource in ?principal);
        // @ts-expect-error - resource slot only allows ?resource
        resource: { op: "in", slot: "?principal" },
      })
    );
  });

  test("rejects resource is variant with invalid in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid for resource constraint): permit(principal, action, resource is Folder in ?principal);
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
        // Cedar (invalid): permit(principal, action, resource ==);
        // @ts-expect-error - resource == requires entity or slot
        resource: { op: "==" },
      })
    );
  });
});
