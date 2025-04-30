import { JSXValueSchema, getValueFromJsonPath } from "./utils";

import { Error } from "./Error";
import { VariableObjectValidated } from "../../../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../../../context/pageCompilerContext";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

type TemplateStringProps = {
  variable: VariableObjectValidated;
};

export const QueryVariable = ({ variable }: TemplateStringProps) => {
  const { src, id, pointer } = variable;
  const { t } = useTranslation();
  const mode = useMode();
  const cacheKey = [id];
  const queryClient = useQueryClient();
  const queryState = queryClient.getQueryState(cacheKey);

  if (queryState === undefined)
    return (
      <Error value={src}>
        {t("direktivPage.error.templateString.query.queryIdNotFound", {
          id,
        })}
      </Error>
    );

  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);
  if (error) {
    return (
      <Error value={src}>
        {t(`direktivPage.error.templateString.query.${error}`)}
      </Error>
    );
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return (
      <Error value={src}>
        {t("direktivPage.error.templateString.query.couldNotStringify")}
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
      {dataParsed.data}
    </span>
  );
};
