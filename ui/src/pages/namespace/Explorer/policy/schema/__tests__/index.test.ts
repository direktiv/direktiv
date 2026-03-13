import { CedarPolicySchema, CedarPolicySchemaType } from "..";
import { createBasePolicy, expectValidPolicy } from "../utils/testutils";
import { describe, expect, test } from "vitest";

describe("Cedar policy schema", () => {
  test("accepts a full policy without annotations", () => {
    /*
      Cedar:
      forbid(
        principal is User in Group::"Admins",
        action in [Action::"readFile", Action::"deleteFile"],
        resource is Folder in Folder::"quarantine"
      )
      when { principal has email && principal.email like "*@example.com" }
      unless { resource getTag "classification" == "public" };
    */
    expectValidPolicy(
      createBasePolicy({
        effect: "forbid",
        principal: {
          op: "is",
          entity_type: "User",
          in: { entity: { type: "Group", id: "Admins" } },
        },
        action: {
          op: "in",
          entities: [
            { type: "Action", id: "readFile" },
            { type: "Action", id: "deleteFile" },
          ],
        },
        resource: {
          op: "is",
          entity_type: "Folder",
          in: { entity: { type: "Folder", id: "quarantine" } },
        },
        conditions: [
          {
            kind: "when",
            body: {
              "&&": {
                left: { has: { left: { Var: "principal" }, attr: "email" } },
                right: {
                  like: {
                    left: {
                      ".": { left: { Var: "principal" }, attr: "email" },
                    },
                    pattern: ["Wildcard", { Literal: "@example.com" }],
                  },
                },
              },
            },
          },
          {
            kind: "unless",
            body: {
              "==": {
                left: {
                  getTag: {
                    left: { Var: "resource" },
                    right: { Value: "classification" },
                  },
                },
                right: { Value: "public" },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects policies without conditions", () => {
    /*
      Cedar (invalid for this schema):
      permit(principal in Group::"Admins", action, resource);
    */
    const invalidPolicy: CedarPolicySchemaType = {
      effect: "permit",
      principal: { op: "in", entity: { type: "Group", id: "Admins" } },
      action: { op: "All" },
      resource: { op: "All" },
      // @ts-expect-error - conditions require a body expression
      conditions: [{ kind: "when" }],
    };

    const runtimeInput = {
      effect: "permit",
      principal: { op: "in", entity: { type: "Group", id: "Admins" } },
      action: { op: "All" },
      resource: { op: "All" },
      conditions: [{ kind: "when" }],
    };

    expect(invalidPolicy).toBeDefined();
    expect(CedarPolicySchema.safeParse(runtimeInput).success).toBe(false);
  });
});
