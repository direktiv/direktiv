import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../testUtils";
import { describe, test } from "vitest";

describe("Cedar principal schema", () => {
  test("accepts principal All", () => {
    expectValidPolicy(createBasePolicy({ principal: { op: "All" } }));
  });

  test("accepts principal == entity", () => {
    const input = createBasePolicy({
      principal: { op: "==", entity: { type: "User", id: "alice" } },
    });

    expectValidPolicy(input);
  });

  test("accepts principal is with in slot", () => {
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

  test("rejects principal == variant with missing entity or slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - principal == requires entity or slot
        principal: { op: "==" },
      })
    );
  });
});
