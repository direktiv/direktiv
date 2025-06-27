import { VariableError } from "./Error";
import { VariableType } from "../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { usePageEditor } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "./utils/useVariableStringResolver";

type VariableProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariableProps) => {
  const { t } = useTranslation();
  const { mode } = usePageEditor();
  const resolveVariableString = useVariableStringResolver();
  const variableString = resolveVariableString(value);

  if (!variableString.success) {
    return (
      <VariableError value={value} errorCode={variableString.error}>
        {t(`direktivPage.error.templateString.${variableString.error}`)}
      </VariableError>
    );
  }

  return (
    <span
      className={twMergeClsx(
        mode === "edit" &&
          "border border-gray-9 bg-gray-4 dark:border-gray-dark-9 dark:bg-gray-dark-4"
      )}
    >
      {variableString.data}
    </span>
  );
};
