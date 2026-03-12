import { CedarPolicySchema, CedarPolicySchemaType } from "..";
import { createBasePolicy, expectValidPolicy } from "../utils/testutils";
import { describe, expect, test } from "vitest";

describe("Cedar policy schema", () => {
  test("accepts a full policy without annotations", () => {
    expectValidPolicy(
      createBasePolicy({
        effect: "forbid",
        principal: { op: "==", entity: { type: "User", id: "alice" } },
        action: { op: "==", entity: { type: "Action", id: "readFile" } },
        resource: { op: "in", entity: { type: "Folder", id: "quarantine" } },
        conditions: [
          {
            kind: "when",
            body: {
              Value: true,
            },
          },
        ],
      })
    );
  });

  test("rejects policies without conditions", () => {
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
