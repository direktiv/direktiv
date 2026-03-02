import { describe, expect, test } from "vitest";

import { CedarPolicySchema, type CedarPolicySchemaType } from "..";

describe("Cedar policy zod schema", () => {
  test("accepts permit effect", () => {
    const input: CedarPolicySchemaType = { effect: "permit" };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("accepts forbid effect", () => {
    const input: CedarPolicySchemaType = { effect: "forbid" };

    expect(CedarPolicySchema.safeParse(input).success).toBe(true);
    expect(CedarPolicySchema.parse(input)).toEqual(input);
  });

  test("rejects unknown effect", () => {
    expect(CedarPolicySchema.safeParse({ effect: "allow" }).success).toBe(
      false
    );
  });
});
