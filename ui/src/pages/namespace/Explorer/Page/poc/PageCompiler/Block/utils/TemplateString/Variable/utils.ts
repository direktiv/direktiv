import {
  VariableNamespaceSchema,
  VariableObject,
  VariableType,
} from "../../../../../schema/primitives/variable";

import { TemplateStringSeparator } from "../../../../../schema/primitives/templateString";
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
 *
 * Note: there is an intentional limitation for the pointer segment here. Arrays
 * will be addressed like this: "data.0.addresses.0.streetName"
 *
 * Meaning that an array position will be treated as if the index of the array were
 * the key in the object. That makes it a lot easier to parse than, for example, the
 * coresponding JavaScript syntax some.data[0].addresses[0].streetName. However, it
 * comes with the limitation that the parser cannot handle JSON data that uses numbers
 * as keys.
 *
 * The syntax also does not allow for keys that have dots in them.
 *
 */
export const parseVariable = (variableString: VariableType): VariableObject => {
  const [namespace, id, ...pointer] = variableString.split(".");
  const parsedNamespace = VariableNamespaceSchema.safeParse(namespace);
  return {
    namespace: parsedNamespace.success ? parsedNamespace.data : undefined,
    id,
    pointer: pointer.length > 0 ? pointer.join(".") : undefined,
  };
};

const AnyObjectSchema = z.object({}).passthrough();
const AnyArraySchema = z.array(z.unknown());
const AnyObjectOrArraySchema = z.union([AnyObjectSchema, AnyArraySchema]);

/**
 * getValueFromJsonPath will get a JSON input and a path to a value in that JSON.
 *
 * It will return a string representation of the value, or a special string if the
 * value is not found or the input is invalid.
 */
type GetValueFromJsonPathSuccess = [unknown, undefined];
type GetValueFromJsonPathFailure = [undefined, "invalidJson" | "invalidPath"];

export const getValueFromJsonPath = (
  json: unknown,
  path: string
): GetValueFromJsonPathSuccess | GetValueFromJsonPathFailure => {
  if (path === "") {
    return [json, undefined];
  }

  if (!AnyObjectOrArraySchema.safeParse(json).success) {
    return [undefined, "invalidJson"];
  }

  const pathSegments = path.split(TemplateStringSeparator);

  let currentSegment: unknown = json;
  for (const segment of pathSegments) {
    const parsed = AnyObjectOrArraySchema.safeParse(currentSegment);
    if (parsed.success && segment in parsed.data) {
      currentSegment = (parsed.data as Record<string, unknown>)[segment];
      continue;
    }
    return [undefined, "invalidPath"];
  }

  return [currentSegment, undefined];
};
