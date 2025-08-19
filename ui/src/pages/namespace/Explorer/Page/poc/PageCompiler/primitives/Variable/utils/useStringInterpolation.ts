import { ResolverFunction } from "./types";
import { parseTemplateString } from ".";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const resolveVariableString = useVariableStringResolver();
  return (input, formVariables) =>
    parseTemplateString(input, (match) => {
      const result = resolveVariableString(match, formVariables);
      if (!result.success) {
        throw new Error(t(`direktivPage.error.templateString.${result.error}`));
      }
      return result.data;
    }).join("");
};
