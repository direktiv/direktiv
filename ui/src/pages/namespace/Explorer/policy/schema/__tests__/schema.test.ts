import { CedarPolicySchema, type CedarPolicySchemaType } from "..";
import { describe, expect, test } from "vitest";

describe("Cedar policy zod schema", () => {
  test("accepts principal All", () => {
    // permit(principal, action, resource);
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts principal == entity", () => {
    // forbid(principal == User::"alice", action, resource);
    const input: CedarPolicySchemaType = {
      effect: "forbid",
      principal: {
        op: "==",
        entity: { type: "User", id: "alice" },
      },
      action: { op: "All" },
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts principal is with in slot", () => {
    // permit(principal is User in ?principal, action, resource);
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: {
        op: "is",
        entity_type: "User",
        in: { slot: "?principal" },
      },
      action: { op: "All" },
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects unknown effect", () => {
    // allow(principal, action, resource);
    expect(
      CedarPolicySchema.safeParse({
        effect: "allow",
        principal: { op: "All" },
        action: { op: "All" },
      }).success
    ).toBe(false);
  });

  test("rejects invalid principal slot", () => {
    // permit(principal == ?resource, action, resource);
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "==", slot: "?resource" },
        action: { op: "All" },
      }).success
    ).toBe(false);
  });

  test("accepts action in entities", () => {
    // permit(principal, action in [Action::"ManageFiles", Action::"readFile"], resource);
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: {
        op: "in",
        entities: [
          { type: "Action", id: "ManageFiles" },
          { type: "Action", id: "readFile" },
        ],
      },
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects invalid action slot", () => {
    // permit(principal, action == ?principal, resource);
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "==", slot: "?principal" },
      }).success
    ).toBe(false);
  });
});
