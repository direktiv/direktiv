import { parseVariable, validateVariable } from "./utils";

import { Error } from "./Error";
import { QueryVariable } from "./Query";
import { VariableType } from "../../../../../schema/primitives/variable";
import { useTranslation } from "react-i18next";

type VariablesProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariablesProps) => {
  const { t } = useTranslation();
  const [variable, error] = validateVariable(parseVariable(value));

  if (error) {
    return (
      <Error value={value}>
        {t(`direktivPage.error.templateString.${error}`)}
      </Error>
    );
  }

  const { namespace } = variable;

  switch (namespace) {
    case "query":
      return <QueryVariable variable={variable} />;
      break;
    default:
      return (
        <Error value={value}>
          {t("direktivPage.error.templateString.namespaceNotImplemented", {
            namespace,
          })}
        </Error>
      );
      break;
  }
};
