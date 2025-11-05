import { describe, expect, test } from "vitest";

import { workflowInputSchema } from "../utils";

describe("workflowInputSchema", () => {
  describe("invalid workflow input", () => {
    test("an empty string is not a valid workflow input", () => {
      expect(workflowInputSchema.safeParse("").success).toEqual(false);
    });

    test("undefined and null is not a valid workflow input", () => {
      expect(workflowInputSchema.safeParse(undefined).success).toEqual(false);
      expect(workflowInputSchema.safeParse(null).success).toEqual(false);
    });

    test("an invalid JSON is not a valid workflow input", () => {
      expect(workflowInputSchema.safeParse("{").success).toEqual(false);
      expect(workflowInputSchema.safeParse("eklmflm").success).toEqual(false);
      expect(workflowInputSchema.safeParse(1).success).toEqual(false);
    });

    test("a JSON without quotes around the key is not a valid workflow input", () => {
      expect(workflowInputSchema.safeParse(`{key:1}`).success).toEqual(false);
      expect(workflowInputSchema.safeParse(`{"key":1}`).success).toEqual(true);
    });

    test("a JSON with trailing comma is not a valid workflow input", () => {
      expect(
        workflowInputSchema.safeParse(`{"key":1,"key":2},`).success
      ).toEqual(false);
      expect(
        workflowInputSchema.safeParse(`{"key":1,"key":2}`).success
      ).toEqual(true);
    });
  });

  describe("invalid workflow input", () => {
    test("an empty JSON is a valid workflow input", () => {
      expect(workflowInputSchema.safeParse("{}").success).toEqual(true);
    });

    test("any valid JSON is a valid workflow input", () => {
      expect(
        workflowInputSchema.safeParse(
          `{"string": "1", "integer": 1, "boolean": true, "array": [1,2,3], "object": {"key": "value"}}`
        ).success
      ).toEqual(true);
    });
  });
});
