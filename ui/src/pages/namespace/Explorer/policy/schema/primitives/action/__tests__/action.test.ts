import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar action schema", () => {
  test("accepts action All operator", () => {
    expectValidPolicy(createBasePolicy({ action: { op: "All" } }));
  });

  test("accepts action == entity", () => {
    const input = createBasePolicy({
      action: { op: "==", entity: { type: "Action", id: "readFile" } },
    });

    expectValidPolicy(input);
  });

  test("accepts action in entity", () => {
    const input = createBasePolicy({
      action: { op: "in", entity: { type: "Action", id: "readOnly" } },
    });

    expectValidPolicy(input);
  });

  test("accepts action in entities", () => {
    const input = createBasePolicy({
      action: {
        op: "in",
        entities: [
          { type: "Action", id: "ManageFiles" },
          { type: "Action", id: "readFile" },
        ],
      },
    });

    expectValidPolicy(input);
  });

  test("rejects invalid action slot", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - action slot only allows ?action
        action: { op: "==", slot: "?principal" },
      })
    );
  });

  test("rejects action in variant with extra keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        action: {
          op: "in",
          entity: { type: "Action", id: "read" },
          entities: [{ type: "Action", id: "write" }],
        },
      })
    );
  });
});
