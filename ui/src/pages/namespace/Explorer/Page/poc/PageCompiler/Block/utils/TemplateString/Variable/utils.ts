import {
  VariableNamespaceSchema,
  VariableObject,
  VariableObjectValidated,
  VariableType,
} from "../../../../../schema/primitives/variable";

import { TemplateStringSeparator } from "../../../../../schema/primitives/templateString";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

/**
 * Regex pattern to match variables enclosed in double curly braces, like {{ variable }}.
 *
 * Explanation:
 * - {{         : Matches the opening double curly braces.
 * - \s*        : Allows optional whitespace after the opening braces.
 * - ([^{}]+?)  : Captures one or more characters that are not { or }.
 * - \s*        : Allows optional whitespace before the closing braces.
 * - }}         : Matches the closing double curly braces literally.
 *
 * The 'g' (global) flag ensures all variable patterns in the string are matched.
 */
export const variablePattern = /{{\s*([^{}]+?)\s*}}/g;

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

type ValidateVariableSuccess = [VariableObjectValidated, undefined];

export type ValidateVariableFailure = [
  undefined,
  "namespaceUndefined" | "idUndefined" | "pointerUndefined",
];

export const validateVariable = (
  variable: VariableObject
): ValidateVariableSuccess | ValidateVariableFailure => {
  const { namespace, id, pointer, src } = variable;

  if (!namespace) return [undefined, "namespaceUndefined"];
  if (!id) return [undefined, "idUndefined"];
  if (!pointer) return [undefined, "pointerUndefined"];

  return [{ src, namespace, id, pointer }, undefined];
};

const AnyObjectSchema = z.object({}).passthrough();
const AnyArraySchema = z.array(z.unknown());
const AnyObjectOrArraySchema = z.union([AnyObjectSchema, AnyArraySchema]);

export type PossibleValues = object | string | number | boolean | null;
type GetValueFromJsonPathSuccess = [PossibleValues, undefined];
export type GetValueFromJsonPathFailure = [
  undefined,
  "invalidJson" | "invalidPath",
];

/**
 * Retrieves a JSON-like input and a path that points to a key in the input.
 *
 * It will return an array of two elements:
 *
 * - The first element is the value at the specified path, or undefined if the
 *   path does not exist or the input is invalid.
 * - The second element is an optional error string if the input or path is invalid.
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
): GetValueFromJsonPathSuccess | GetValueFromJsonPathFailure => {
  const jsonParsed = AnyObjectOrArraySchema.safeParse(json);
  if (!jsonParsed.success) {
    return [undefined, "invalidJson"];
  }

  if (path === "") {
    return [jsonParsed.data, undefined];
  }

  const pathSegments = path.split(TemplateStringSeparator);

  let returnValue: PossibleValues = jsonParsed.data;

  for (const segment of pathSegments) {
    const returnValueParsed = AnyObjectOrArraySchema.safeParse(returnValue);
    if (returnValueParsed.success && segment in returnValueParsed.data) {
      returnValue = (returnValueParsed.data as Record<string, unknown>)[
        segment
      ] as PossibleValues;
      continue;
    }

    return [undefined, "invalidPath"];
  }

  return [returnValue, undefined];
};

type UseVariableSuccess = [PossibleValues, undefined];
type QueryFailure = [undefined, "queryIdNotFound" | "couldNotStringify"];

type UseVariableFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | QueryFailure;

export const useVariable = (
  value: VariableType
): UseVariableSuccess | UseVariableFailure => {
  const queryClient = useQueryClient();
  const [variableObject, validationError] = validateVariable(
    parseVariable(value)
  );

  if (validationError) {
    return [undefined, validationError];
  }

  const { id, pointer } = variableObject;
  const cacheKey = [id];
  const queryState = queryClient.getQueryState(cacheKey);

  if (queryState === undefined) {
    return [undefined, "queryIdNotFound"];
  }

  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);

  if (error) {
    return [undefined, error];
  }

  return [data, undefined];
};

export const JSXValueSchema = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
  z.undefined(),
]);

export type JSXValueType = z.infer<typeof JSXValueSchema>;

type UseVariableJSXSuccess = [JSXValueType, undefined];

type UseVariableJSXFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | QueryFailure;

export const useVariableJSX = (
  value: VariableType
): UseVariableJSXSuccess | UseVariableJSXFailure => {
  const [data, error] = useVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "couldNotStringify"];
  }

  return [dataParsed.data, undefined];
};
