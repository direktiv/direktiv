import {
  complexValidationAsFirstState,
  noValidateState,
  validationAsFirstState,
  validationAsSecondState,
} from "./workflowTemplates";
import { describe, expect, test } from "vitest";
import { getValidationSchemaFromYaml, workflowInputSchema } from "../utils";

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

describe("getValidationSchemaFromYaml", () => {
  describe("valid yaml input", () => {
    test("a workflow with a validation state as the first state will return the equivalent JSONschema", () => {
      expect(getValidationSchemaFromYaml(validationAsFirstState)).toEqual({
        type: "object",
        properties: { email: { type: "string", format: "email" } },
      });
    });

    test("it supports required fields and a title inside the validation step and reflect them in the JSONschema", () => {
      expect(
        getValidationSchemaFromYaml(complexValidationAsFirstState)
      ).toEqual({
        type: "object",
        title: "A registration form",
        required: ["firstName", "lastName"],
        properties: {
          age: {
            title: "Age",
            type: "integer",
          },
          firstName: {
            title: "First name",
            type: "string",
          },
          bio: {
            title: "Bio",
            type: "string",
          },
          lastName: {
            title: "Last name",
            type: "string",
          },
          password: {
            title: "Password",
            type: "string",
          },
        },
      });
    });

    test("a workflow with a validation state NOT as the first state will return null", () => {
      expect(getValidationSchemaFromYaml(validationAsSecondState)).toEqual(
        null
      );
    });

    test("a workflow with no validation state will return null", () => {
      expect(getValidationSchemaFromYaml(noValidateState)).toEqual(null);
    });
  });

  describe("it returns null for any invalid yaml input", () => {
    test("undefined input", () => {
      expect(getValidationSchemaFromYaml(undefined)).toBe(null);
    });

    test("empty string input", () => {
      expect(getValidationSchemaFromYaml("")).toBe(null);
    });

    test("random unexpected string input", () => {
      expect(getValidationSchemaFromYaml("some random string")).toBe(null);
      expect(getValidationSchemaFromYaml("\n\n\n\nsome random string")).toBe(
        null
      );
      expect(getValidationSchemaFromYaml("🙃")).toBe(null);
    });
  });
});
