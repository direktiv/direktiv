import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

describe("Cedar action schema", () => {
  test("accepts action All operator", () => {
    // Cedar: permit(principal, action, resource);
    expectValidPolicy(createBasePolicy({ action: { op: "All" } }));
  });

  test("accepts action == entity", () => {
    // Cedar: permit(principal, action == Action::"readFile", resource);
    const input = createBasePolicy({
      action: { op: "==", entity: { type: "Action", id: "readFile" } },
    });

    expectValidPolicy(input);
  });

  test("accepts action in entity", () => {
    // Cedar: permit(principal, action in Action::"readOnly", resource);
    const input = createBasePolicy({
      action: { op: "in", entity: { type: "Action", id: "readOnly" } },
    });

    expectValidPolicy(input);
  });

  test("accepts action in entities", () => {
    // Cedar: permit(principal, action in [Action::"ManageFiles", Action::"readFile"], resource);
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
        // Cedar (invalid for this schema): permit(principal, action == ?principal, resource);
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
          // @ts-expect-error - action in variants are mutually exclusive
          entities: [{ type: "Action", id: "write" }],
        },
      })
    );
  });
});
