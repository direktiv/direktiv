import { JsonPathError, ValidateVariableError } from "./errors";
import {
  TemplateStringSeparator,
  TemplateStringType,
} from "../../../../schema/primitives/templateString";
import {
  VariableNamespaceSchema,
  VariableObject,
  VariableObjectValidated,
  VariableType,
  localVariableNamespace,
} from "../../../../schema/primitives/variable";

import { ValidationResult } from "./types";
import { z } from "zod";

/**
 * Regex pattern to match variables enclosed in double curly braces, like {{variable}}.
 *
 * Explanation:
 * - {{         : Matches the opening double curly braces.
 * - \s*        : Allows optional whitespace after the opening braces.
 * - ([^\s{}]+) : Captures one or more characters that are not whitespace, {, or }.
 * - \s*        : Allows optional whitespace before the closing braces.
 * - }}         : Matches the closing double curly braces literally.
 *
 * The 'g' (global) flag ensures all variable patterns in the string are matched.
 */
export const variablePattern = /{{\s*([^\s{}]+)\s*}}/g;

/**
 * Splits a template string into an array of text fragments and processed variables.
 *
 * The function processes a string containing variables in the format {{variable}}
 * and returns an array where:
 * - text fragments are left untouched
 * - variables are processed through the provided callback function
 * - empty strings are included to preserve positional information
 *
 * Example: "Hello {{name}}!" with callback (match) => `[${match}]` returns:
 * ["Hello ", "[name]", "!"]
 */
export const parseTemplateString = <T>(
  value: TemplateStringType,
  onMatch: (match: string, index: number) => T
): (string | T)[] => {
  const fragments = value.split(variablePattern);

  return fragments.map((fragment, index) => {
    const isVariable = index % 2 === 1;

    if (isVariable) {
      return onMatch(fragment, index);
    }

    return fragment;
  });
};

/**
 * Parses a variable string into its individual components.
 *
 * Example: "query.company-list.data.0.name" will be parsed into:
 * {
 *   namespace: "query",
 *   id: "company-list",
 *   pointer: "data.0.name"
 * }
 */
export const parseVariable = (variableString: VariableType): VariableObject => {
  const [namespace, id, ...pointer] = variableString.split(".");
  const parsedNamespace = VariableNamespaceSchema.safeParse(namespace);
  return {
    src: variableString,
    namespace: parsedNamespace.success ? parsedNamespace.data : undefined,
    id,
    pointer: pointer.length > 0 ? pointer.join(".") : undefined,
  };
};

export const validateVariable = (
  variable: VariableObject
): ValidationResult<VariableObjectValidated, ValidateVariableError> => {
  const { namespace, id, pointer, src } = variable;

  if (!namespace) return { success: false, error: "namespaceInvalid" };
  if (!id) return { success: false, error: "idUndefined" };

  if (namespace === localVariableNamespace) {
    return { success: true, data: { src, namespace, id } };
  }

  if (!pointer) return { success: false, error: "pointerUndefined" };
  return { success: true, data: { src, namespace, id, pointer } };
};

const AnyObjectSchema = z.object({}).passthrough();
const AnyArraySchema = z.array(z.unknown());
const AnyObjectOrArraySchema = z.union([AnyObjectSchema, AnyArraySchema]);

export type JsonValueType = object | string | number | boolean | null;

/**
 * Retrieves a JSON-like input and a path that points to a key in the input.
 *
 * It will return a Result object with either:
 * - The value at the specified path on success
 * - An error string if the input or path is invalid
 *
 * Path notation:
 * - Arrays are addressed as if their indices are object keys, e.g.,
 *   "data.0.addresses.0.streetName".
 * - This simplifies parsing compared to standard JavaScript syntax like
 *   `data[0].addresses[0].streetName`.
 *
 * Limitations:
 * - If a path contains numbered keys as explained above, it is unclear whether the
 *   numbers are indices or keys
 * - Keys containing dots (".") are not supported.
 */
export const getValueFromJsonPath = (
  json: unknown,
  path: string
): ValidationResult<JsonValueType, JsonPathError> => {
  const jsonParsed = AnyObjectOrArraySchema.safeParse(json);
  if (!jsonParsed.success) {
    return { success: false, error: "invalidJson" };
  }

  if (path === "") {
    return { success: true, data: jsonParsed.data };
  }

  const pathSegments = path.split(TemplateStringSeparator);

  let returnValue: JsonValueType = jsonParsed.data;

  for (const segment of pathSegments) {
    const returnValueParsed = AnyObjectOrArraySchema.safeParse(returnValue);
    if (returnValueParsed.success && segment in returnValueParsed.data) {
      returnValue = (returnValueParsed.data as Record<string, unknown>)[
        segment
      ] as JsonValueType;
      continue;
    }

    return { success: false, error: "invalidPath" };
  }

  return { success: true, data: returnValue };
};
