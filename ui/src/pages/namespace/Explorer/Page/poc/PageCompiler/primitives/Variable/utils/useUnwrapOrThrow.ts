import { AllVariableErrors } from "./errors";
import { ValidationResult } from "./types";
import { useTranslation } from "react-i18next";

/**
 * Hook that returns a function to unwrap validation results or
 * throw translated errors.
 *
 * returns a function that takes a ValidationResult and variable
 * name, returning the unwrapped data if successful or throwing
 * a translated error if validation failed.
 */
export const useUnwrapOrThrow = () => {
  const { t } = useTranslation();
  return <T, E extends AllVariableErrors>(
    result: ValidationResult<T, E>,
    variable: string
  ): T => {
    if (!result.success) {
      throw new Error(
        t(`direktivPage.error.templateString.${result.error}`, {
          variable,
        })
      );
    }
    return result.data;
  };
};
