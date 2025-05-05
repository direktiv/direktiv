import { Error } from "./Error";
import { VariableType } from "../../../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";
import { useVariableJSX } from "./utils/useVariableJSX";

type TemplateStringProps = {
  variable: VariableType;
};

export const Variable = ({ variable }: TemplateStringProps) => {
  const { t } = useTranslation();
  const mode = useMode();
  const [variableContent, error] = useVariableJSX(variable);

  if (error) {
    return (
      <Error value={variable}>
        {t(`direktivPage.error.templateString.${error}`)}
      </Error>
    );
  }

  return (
    <span
      className={twMergeClsx(
        mode === "preview" &&
          "border border-gray-9 bg-gray-4 dark:bg-gray-dark-4 dark:border-gray-dark-9"
      )}
    >
      {variableContent}
    </span>
  );
};
