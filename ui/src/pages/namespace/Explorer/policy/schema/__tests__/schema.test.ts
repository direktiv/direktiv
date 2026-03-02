import { CedarPolicySchema, type CedarPolicySchemaType } from "..";
import { describe, expect, test } from "vitest";

describe("Cedar policy zod schema", () => {
  test("accepts principal All", () => {
    // permit(principal, action, resource);
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [],
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
      resource: { op: "All" },
      conditions: [],
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
      resource: { op: "All" },
      conditions: [],
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
        resource: { op: "All" },
        conditions: [],
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
        resource: { op: "All" },
        conditions: [],
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
      resource: { op: "All" },
      conditions: [],
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
        resource: { op: "All" },
        conditions: [],
      }).success
    ).toBe(false);
  });

  test("accepts resource is with in slot", () => {
    // permit(principal, action, resource is Folder in ?resource);
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: {
        op: "is",
        entity_type: "Folder",
        in: { slot: "?resource" },
      },
      conditions: [],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects invalid resource slot", () => {
    // permit(principal, action, resource == ?principal);
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "==", slot: "?principal" },
        conditions: [],
      }).success
    ).toBe(false);
  });

  test("accepts when condition", () => {
    // permit(principal, action, resource) when { true };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            Value: true,
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts condition with Var JsonExpr", () => {
    // permit(principal, action, resource) when { context };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            Var: "context",
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects invalid Var JsonExpr value", () => {
    // permit(principal, action, resource) when { actor };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [{ kind: "when", body: { Var: "actor" } }],
      }).success
    ).toBe(false);
  });

  test("accepts condition with Slot JsonExpr", () => {
    // permit(principal, action, resource) when { ?principal };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            Slot: "?principal",
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects invalid Slot JsonExpr value", () => {
    // permit(principal, action, resource) when { ?action };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [{ kind: "when", body: { Slot: "?action" } }],
      }).success
    ).toBe(false);
  });

  test("accepts condition with Unknown JsonExpr", () => {
    // permit(principal, action, resource) when { Unknown("x") };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            Unknown: { name: "x" },
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects Unknown JsonExpr with multiple keys", () => {
    // permit(principal, action, resource) when { Unknown("x", "y") };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [
          {
            kind: "when",
            body: {
              Unknown: { a: "x", b: "y" },
            },
          },
        ],
      }).success
    ).toBe(false);
  });

  test("accepts condition with unary ! JsonExpr", () => {
    // permit(principal, action, resource) when { !context };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            "!": {
              arg: {
                Var: "context",
              },
            },
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts condition with unary neg JsonExpr", () => {
    // permit(principal, action, resource) when { -1 };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            neg: {
              arg: {
                Value: 1,
              },
            },
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects unary JsonExpr without arg", () => {
    // permit(principal, action, resource) when { ! };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [{ kind: "when", body: { "!": {} } }],
      }).success
    ).toBe(false);
  });

  test("rejects invalid condition kind", () => {
    // permit(principal, action, resource) iff { true };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [{ kind: "iff", body: { Value: true } }],
      }).success
    ).toBe(false);
  });

  test("accepts condition with binary == JsonExpr", () => {
    // permit(principal, action, resource) when { context == "1.3" };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: {
                  Var: "context",
                },
                right: {
                  Value: "1.3",
                },
              },
            },
          },
        ],
      }).success
    ).toBe(true);
  });

  test("rejects binary JsonExpr without right operand", () => {
    // permit(principal, action, resource) when { context == };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: {
                  Var: "context",
                },
              },
            },
          },
        ],
      }).success
    ).toBe(false);
  });

  test("accepts condition with attribute . JsonExpr", () => {
    // permit(principal, action, resource) when { context.tls_version };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            ".": {
              left: { Var: "context" },
              attr: "tls_version",
            },
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts condition with attribute has JsonExpr", () => {
    // permit(principal, action, resource) when { principal has "email" };
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [
        {
          kind: "when",
          body: {
            has: {
              left: { Var: "principal" },
              attr: "email",
            },
          },
        },
      ],
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects attribute JsonExpr without attr", () => {
    // permit(principal, action, resource) when { context. };
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [
          {
            kind: "when",
            body: {
              ".": {
                left: { Var: "context" },
              },
            },
          },
        ],
      }).success
    ).toBe(false);
  });

  test("accepts annotations with string and null", () => {
    // @shadow_mode, @reason("temporary block")
    const input: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "All" },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [],
      annotations: {
        shadow_mode: null,
        reason: "temporary block",
      },
    };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects invalid annotation value type", () => {
    // @priority(10)
    expect(
      CedarPolicySchema.safeParse({
        effect: "permit",
        principal: { op: "All" },
        action: { op: "All" },
        resource: { op: "All" },
        conditions: [],
        annotations: {
          priority: 10,
        },
      }).success
    ).toBe(false);
  });
});
