import { describe, expect, test } from "vitest";
import {
  noValidateState,
  validationAsFirstState,
  validationAsSecondState,
} from "./workflowTemplates";

import { getValidationSchemaFromYaml } from "../utils";

describe("getValidationSchemaFromYaml", () => {
  describe("valid yaml input", () => {
    test("a workflow with a validation state as the first state will return the JSONschema", () => {
      expect(getValidationSchemaFromYaml(validationAsFirstState)).toEqual({
        type: "object",
        properties: { email: { type: "string", format: "email" } },
      });
    });

    test("a workflow with a validation state NOT as the first state will return null", () => {
      expect(getValidationSchemaFromYaml(validationAsSecondState)).toEqual(
        null
      );
    });

    test("a workflow with not validation state will return null", () => {
      expect(getValidationSchemaFromYaml(noValidateState)).toEqual(null);
    });
  });

  describe("it returns null when the for any invalid yaml input", () => {
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
