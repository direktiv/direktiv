import {
  complexValidationAsFirstState,
  noValidateState,
  validationAsFirstState,
  validationAsSecondState,
} from "./workflowTemplates";
import { describe, expect, test } from "vitest";

import { getValidationSchemaFromYaml } from "../utils";

describe("getValidationSchemaFromYaml", () => {
  describe("valid yaml input", () => {
    test("a workflow with a validation state as the first state will return the JSONschema", () => {
      expect(getValidationSchemaFromYaml(validationAsFirstState)).toEqual({
        type: "object",
        properties: { email: { type: "string", format: "email" } },
      });
    });

    test("it supports required fields and a title in the JSONschema", () => {
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
