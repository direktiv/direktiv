import { describe, test } from "vitest";

import {
  expectInvalidPolicySet,
  expectValidPolicySet,
} from "../../../testUtils";
import type { CedarPolicySetSchemaType } from "../../../index";

describe("Cedar policy set schema", () => {
  test("accepts staticPolicies, templates, and templateLinks", () => {
    // policy set with one static policy and one template link value for ?resource
    const input: CedarPolicySetSchemaType = {
      staticPolicies: {
        policy0: {
          effect: "permit",
          principal: { op: "All" },
          action: { op: "All" },
          resource: { op: "All" },
          conditions: [],
        },
      },
      templates: {
        template0: {
          effect: "forbid",
          principal: { op: "All" },
          action: { op: "All" },
          resource: { op: "in", slot: "?resource" },
          conditions: [],
        },
      },
      templateLinks: [
        {
          templateId: "template0",
          newId: "policy1",
          values: {
            "?resource": { type: "Folder", id: "def" },
          },
        },
      ],
    };

    expectValidPolicySet(input);
  });

  test("rejects template link with invalid slot key", () => {
    // template link values can only use ?principal or ?resource
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "template0",
          newId: "policy1",
          values: {
            // @ts-expect-error - only ?principal and ?resource are allowed
            "?action": { type: "Action", id: "read" },
          },
        },
      ],
    };

    expectInvalidPolicySet(input);
  });
});
