import { AllVariableErrors } from "./errors";
import { ValidationResult } from "./types";
import { useTranslation } from "react-i18next";

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
