import {
  VariableNamespaceSchema,
  VariableObject,
  VariableType,
} from "../../../../../schema/primitives/variable";

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

// TODO: add unit tests
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
    namespace: parsedNamespace.success ? parsedNamespace.data : undefined,
    id,
    pointer: pointer.length > 0 ? pointer.join(".") : undefined,
  };
};
