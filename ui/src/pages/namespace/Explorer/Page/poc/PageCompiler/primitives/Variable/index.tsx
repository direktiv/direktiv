import { VariableError } from "./Error";
import { VariableType } from "../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { usePageEditor } from "../../context/pageCompilerContext";
import { useResolveVariableString } from "./utils/useResolveVariableString";
import { useTranslation } from "react-i18next";

type VariableProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariableProps) => {
  const { t } = useTranslation();
  const { mode } = usePageEditor();
  const variableJSX = useResolveVariableString(value);

  if (!variableJSX.success) {
    return (
      <VariableError value={value} errorCode={variableJSX.error}>
        {t(`direktivPage.error.templateString.${variableJSX.error}`)}
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
      {variableJSX.data}
    </span>
  );
};
