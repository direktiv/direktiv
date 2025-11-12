import { ResolverFunction } from "./types";
import { parseTemplateString } from ".";
import { useUnwrapOrThrow } from "./useUnwrapOrThrow";
import { useVariableStringResolver } from "./useVariableStringResolver";

/**
 * A hook that returns a function that enables you to analyze a string
 * for variable strings and replace them with their resolved value from
 * the React context.
 *
 * Example:
 *
 * const string = "company-id-{{query.company-list.data.0.id}}";
 * const interpolateString = useStringInterpolation();
 *
 * console.log(interpolateString(string)); // "company-id-apple"
 *
 */
export const useStringInterpolation = (): ResolverFunction<string> => {
  const unwrapOrThrow = useUnwrapOrThrow();
  const resolveVariableString = useVariableStringResolver();
  return (input, localVariables) =>
    parseTemplateString(input, (match) => {
      const result = resolveVariableString(match, localVariables);
      return unwrapOrThrow(result, match);
    }).join("");
};
