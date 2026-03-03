import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "./utils";
import { describe, test } from "vitest";

describe("Cedar policy schema", () => {
  test("accepts principal All", () => {
    // permit(principal, action, resource);
    expectValidPolicy(createBasePolicy());
  });

  test("accepts principal == entity", () => {
    // forbid(principal == User::"alice", action, resource);
    const input = createBasePolicy({
      effect: "forbid",
      principal: { op: "==", entity: { type: "User", id: "alice" } },
    });

    expectValidPolicy(input);
  });

  test("accepts principal is with in slot", () => {
    // permit(principal is User in ?principal, action, resource);
    const input = createBasePolicy({
      principal: { op: "is", entity_type: "User", in: { slot: "?principal" } },
    });

    expectValidPolicy(input);
  });

  test("rejects unknown effect", () => {
    // allow(principal, action, resource);
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - only permit/forbid are allowed
        effect: "allow",
      })
    );
  });

  test("rejects invalid principal slot", () => {
    // permit(principal == ?resource, action, resource);
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - principal slot only allows ?principal
        principal: { op: "==", slot: "?resource" },
      })
    );
  });

  test("accepts action in entities", () => {
    // permit(principal, action in [Action::"ManageFiles", Action::"readFile"], resource);
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
    // permit(principal, action == ?principal, resource);
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - action slot only allows ?action
        action: { op: "==", slot: "?principal" },
      })
    );
  });

  test("accepts resource is with in slot", () => {
    // permit(principal, action, resource is Folder in ?resource);
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
    // permit(principal, action, resource == ?principal);
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - resource slot only allows ?resource
        resource: { op: "==", slot: "?principal" },
      })
    );
  });

  test("accepts annotations with string and null", () => {
    // @shadow_mode, @reason("temporary block")
    const input = createBasePolicy({
      annotations: {
        shadow_mode: null,
        reason: "temporary block",
      },
    });

    expectValidPolicy(input);
  });

  test("rejects invalid annotation value type", () => {
    // @priority(10)
    expectInvalidPolicy(
      createBasePolicy({
        annotations: {
          // @ts-expect-error - annotation values must be string or null
          priority: 10,
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
