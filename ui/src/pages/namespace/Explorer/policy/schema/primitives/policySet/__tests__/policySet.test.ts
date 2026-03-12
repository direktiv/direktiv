import {
  createBasePolicy,
  expectInvalidPolicySet,
  expectValidPolicySet,
} from "../../../utils/testutils";
import { describe, test } from "vitest";

import type { CedarPolicySetSchemaType } from "../../../index";

describe("Cedar policy set schema", () => {
  test("accepts empty policy set", () => {
    const input: CedarPolicySetSchemaType = {};

    expectValidPolicySet(input);
  });

  test("accepts only staticPolicies", () => {
    const input: CedarPolicySetSchemaType = {
      staticPolicies: {
        policy0: createBasePolicy(),
      },
    };

    expectValidPolicySet(input);
  });

  test("accepts only templates", () => {
    const input: CedarPolicySetSchemaType = {
      templates: {
        template0: createBasePolicy({
          effect: "forbid",
          resource: { op: "in", slot: "?resource" },
        }),
      },
    };

    expectValidPolicySet(input);
  });

  test("accepts only templateLinks", () => {
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "template0",
          newId: "policy0",
          values: {
            "?principal": { type: "User", id: "alice" },
            "?resource": { type: "Folder", id: "def" },
          },
        },
      ],
    };

    expectValidPolicySet(input);
  });

  test("accepts staticPolicies, templates, and templateLinks", () => {
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

  test("rejects template link with empty templateId", () => {
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "",
          newId: "policy1",
          values: {
            "?resource": { type: "Folder", id: "def" },
          },
        },
      ],
    };

    expectInvalidPolicySet(input);
  });

  test("rejects template link with empty newId", () => {
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "template0",
          newId: "",
          values: {
            "?resource": { type: "Folder", id: "def" },
          },
        },
      ],
    };

    expectInvalidPolicySet(input);
  });

  test("rejects template link with invalid entity value", () => {
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "template0",
          newId: "policy1",
          values: {
            // @ts-expect-error - entity requires both type and id
            "?resource": { type: "Folder" },
          },
        },
      ],
    };

    expectInvalidPolicySet(input);
  });

  test("rejects policy set with unknown top-level key", () => {
    const input: CedarPolicySetSchemaType = {
      staticPolicies: {
        policy0: createBasePolicy(),
      },
      // @ts-expect-error - unknown key
      extra: true,
    };

    expectInvalidPolicySet(input);
  });

  test("rejects template link with unknown key", () => {
    const input: CedarPolicySetSchemaType = {
      templateLinks: [
        {
          templateId: "template0",
          newId: "policy1",
          values: {
            "?resource": { type: "Folder", id: "def" },
          },
          // @ts-expect-error - unknown key
          extra: true,
        },
      ],
    };

    expectInvalidPolicySet(input);
  });

  test("rejects invalid policy in staticPolicies", () => {
    const input: CedarPolicySetSchemaType = {
      staticPolicies: {
        policy0: createBasePolicy({
          // @ts-expect-error - principal slot only allows ?principal
          principal: { op: "==", slot: "?resource" },
        }),
      },
    };

    expectInvalidPolicySet(input);
  });

  test("rejects invalid policy in templates", () => {
    const input: CedarPolicySetSchemaType = {
      templates: {
        template0: createBasePolicy({
          // @ts-expect-error - resource == requires entity or slot
          resource: { op: "==" },
        }),
      },
    };

    expectInvalidPolicySet(input);
  });
});
