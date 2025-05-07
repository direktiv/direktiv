import { Error } from "./Error";
import { VariableType } from "../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../context/pageCompilerContext";
import { useResolveVariableJSX } from "./utils/useResolveVariableJSX";
import { useTranslation } from "react-i18next";

type TemplateStringProps = {
  value: VariableType;
};

export const Variable = ({ value }: TemplateStringProps) => {
  const { t } = useTranslation();
  const mode = useMode();
  const [variableContent, error] = useResolveVariableJSX(value);

  if (error) {
    return (
      <Error value={value}>
        {t(`direktivPage.error.templateString.${error}`)}
      </Error>
    );
  }

  return (
    <span
      className={twMergeClsx(
        mode === "inspect" &&
          "border border-gray-9 bg-gray-4 dark:bg-gray-dark-4 dark:border-gray-dark-9"
      )}
    >
      {variableContent}
    </span>
  );
};
