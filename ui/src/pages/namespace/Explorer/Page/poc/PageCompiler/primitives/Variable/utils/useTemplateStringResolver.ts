import { TemplateStringType } from "../../../../schema/primitives/templateString";
import { parseTemplateString } from ".";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "./useVariableStringResolver";

/**
 * A hook that processes a template string and returns the resolved string
 * with variables replaced by their actual values from the React context.
 *
 * This is the string equivalent of the TemplateString JSX component.
 *
 * Example:
 *
 * const templateString = "Hello {{user.name}}, your order {{order.id}} is ready!";
 * const useTemplate = useTemplateString();
 * const resolvedString = useTemplate(templateString);
 *
 * console.log(resolvedString); // "Hello John, your order 12345 is ready!"
 */
export const useTemplateStringResolver = () => {
  const { t } = useTranslation();
  const resolveVariableString = useVariableStringResolver();

  return (value: TemplateStringType): string => {
    const templateFragments = parseTemplateString(value, (match) => {
      const result = resolveVariableString(match);
      if (!result.success) {
        throw new Error(t(`direktivPage.error.templateString.${result.error}`));
      }
      return String(result.data);
    });

    return templateFragments.join("");
  };
};
