import {
  VariableNamespaceSchema,
  VariableObject,
  VariableType,
} from "../../../../../schema/primitives/variable";

import { TemplateStringSeparator } from "../../../../../schema/primitives/templateString";

/**
 * Regex to match variables enclosed in double curly braces, like {{ variable }}.
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

export const getObjectValueByPath = (
  obj: unknown,
  path: string
): string | undefined => {
  if (!obj || !path || typeof obj !== "object") {
    return undefined;
  }

  const pathParts = path.split(TemplateStringSeparator);
  let current = obj;

  for (const part of pathParts) {
    if (
      current &&
      typeof current === "object" &&
      current !== null &&
      part in current
    ) {
      current = (current as Record<string, unknown>)[part];
    } else {
      return undefined; // Path not found
    }
  }

  return current;
};
