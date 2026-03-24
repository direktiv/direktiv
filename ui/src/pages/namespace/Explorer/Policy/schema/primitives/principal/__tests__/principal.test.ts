import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar principal schema", () => {
  test("accepts principal All operator", () => {
    // Cedar: permit(principal, action, resource);
    expectValidPolicy(createBasePolicy({ principal: { op: "All" } }));
  });

  test("accepts principal == entity", () => {
    // Cedar: permit(principal == User::"alice", action, resource);
    const input = createBasePolicy({
      principal: { op: "==", entity: { type: "User", id: "alice" } },
    });
    expectValidPolicy(input);
  });

  test("accepts principal == slot", () => {
    // Cedar template: permit(principal == ?principal, action, resource);
    const input = createBasePolicy({
      principal: { op: "==", slot: "?principal" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal in entity", () => {
    // Cedar: permit(principal in Group::"Admins", action, resource);
    const input = createBasePolicy({
      principal: { op: "in", entity: { type: "Group", id: "Admins" } },
    });
    expectValidPolicy(input);
  });

  test("accepts principal in slot", () => {
    // Cedar template: permit(principal in ?principal, action, resource);
    const input = createBasePolicy({
      principal: { op: "in", slot: "?principal" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal is entity type", () => {
    // Cedar: permit(principal is User, action, resource);
    const input = createBasePolicy({
      principal: { op: "is", entity_type: "User" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal is in entity", () => {
    // Cedar: permit(principal is User in Group::"Admins", action, resource);
    const input = createBasePolicy({
      principal: {
        op: "is",
        entity_type: "User",
        in: { entity: { type: "Group", id: "Admins" } },
      },
    });
    expectValidPolicy(input);
  });

  test("accepts principal is in slot", () => {
    // Cedar template: permit(principal is User in ?principal, action, resource);
    const input = createBasePolicy({
      principal: { op: "is", entity_type: "User", in: { slot: "?principal" } },
    });
    expectValidPolicy(input);
  });

  test("rejects invalid principal slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid for principal constraint): permit(principal == ?resource, action, resource);
        // @ts-expect-error - principal slot only allows ?principal
        principal: { op: "==", slot: "?resource" },
      })
    );
  });

  test("rejects invalid principal in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid for principal constraint): permit(principal in ?resource, action, resource);
        // @ts-expect-error - principal slot only allows ?principal
        principal: { op: "in", slot: "?resource" },
      })
    );
  });

  test("rejects principal is variant with invalid in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid for principal constraint): permit(principal is User in ?resource, action, resource);
        // @ts-expect-error - principal is/in slot only allows ?principal
        principal: {
          op: "is",
          entity_type: "User",
          in: { slot: "?resource" },
        },
      })
    );
  });

  test("rejects principal == variant with missing entity or slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // Cedar (invalid): permit(principal ==, action, resource);
        // @ts-expect-error - principal == requires entity or slot
        principal: { op: "==" },
      })
    );
  });
});
