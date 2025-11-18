import { VariableError } from "./Error";
import { VariableType } from "../../../schema/primitives/variable";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "./utils/useVariableStringResolver";

type VariableProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariableProps) => {
  const { t } = useTranslation();
  const resolveVariableString = useVariableStringResolver();
  const variableString = resolveVariableString(value);

  if (!variableString.success) {
    return (
      <VariableError value={value} errorCode={variableString.error}>
        {t(`direktivPage.error.templateString.${variableString.error}`, {
          variable: value,
        })}
      </VariableError>
    );
  }

  return <span>{variableString.data}</span>;
};
