import {
  CedarPolicySchema,
  type CedarPolicySchemaType,
  CedarPolicySetSchema,
  type CedarPolicySetSchemaType,
} from "..";
import { expect } from "vitest";

export const createBasePolicy = (
  overrides: Partial<CedarPolicySchemaType> = {}
): CedarPolicySchemaType => ({
  effect: "permit",
  principal: { op: "All" },
  action: { op: "All" },
  resource: { op: "All" },
  conditions: [],
  ...overrides,
});

export const expectValidPolicy = (input: CedarPolicySchemaType) => {
  expect(CedarPolicySchema.safeParse(input).success).toBe(true);
  expect(CedarPolicySchema.parse(input)).toEqual(input);
};

export const expectInvalidPolicy = (input: CedarPolicySetSchemaType) => {
  expect(CedarPolicySchema.safeParse(input).success).toBe(false);
};

export const expectValidPolicySet = (input: CedarPolicySetSchemaType) => {
  expect(CedarPolicySetSchema.safeParse(input).success).toBe(true);
  expect(CedarPolicySetSchema.parse(input)).toEqual(input);
};

export const expectInvalidPolicySet = (input: CedarPolicySetSchemaType) => {
  expect(CedarPolicySetSchema.safeParse(input).success).toBe(false);
};
