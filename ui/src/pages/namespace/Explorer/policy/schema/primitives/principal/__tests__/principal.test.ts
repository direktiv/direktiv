import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar principal schema", () => {
  test("accepts principal All operator", () => {
    expectValidPolicy(createBasePolicy({ principal: { op: "All" } }));
  });

  test("accepts principal == entity", () => {
    const input = createBasePolicy({
      principal: { op: "==", entity: { type: "User", id: "alice" } },
    });
    expectValidPolicy(input);
  });

  test("accepts principal == slot", () => {
    const input = createBasePolicy({
      principal: { op: "==", slot: "?principal" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal in entity", () => {
    const input = createBasePolicy({
      principal: { op: "in", entity: { type: "Group", id: "Admins" } },
    });
    expectValidPolicy(input);
  });

  test("accepts principal in slot", () => {
    const input = createBasePolicy({
      principal: { op: "in", slot: "?principal" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal is entity type", () => {
    const input = createBasePolicy({
      principal: { op: "is", entity_type: "User" },
    });
    expectValidPolicy(input);
  });

  test("accepts principal is in entity", () => {
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
    const input = createBasePolicy({
      principal: { op: "is", entity_type: "User", in: { slot: "?principal" } },
    });
    expectValidPolicy(input);
  });

  test("rejects invalid principal slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - principal slot only allows ?principal
        principal: { op: "==", slot: "?resource" },
      })
    );
  });

  test("rejects invalid principal in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - principal slot only allows ?principal
        principal: { op: "in", slot: "?resource" },
      })
    );
  });

  test("rejects principal is variant with invalid in slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
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
        // @ts-expect-error - principal == requires entity or slot
        principal: { op: "==" },
      })
    );
  });
});
