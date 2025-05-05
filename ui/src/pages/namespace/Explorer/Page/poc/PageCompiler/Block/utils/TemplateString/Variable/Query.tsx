import {
  GetValueFromJsonPathFailure,
  JSXValueSchema,
  JSXValueType,
  getValueFromJsonPath,
} from "./utils";

import { Error } from "./Error";
import { VariableObjectValidated } from "../../../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../../../context/pageCompilerContext";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

type TemplateStringProps = {
  variable: VariableObjectValidated;
};

type UseVariableSuccess = [JSXValueType, undefined];
type QueryFailure = [undefined, "queryIdNotFound" | "couldNotStringify"];
type UseVariableFailure = GetValueFromJsonPathFailure | QueryFailure;

export const useQueryVariable = (
  variable: VariableObjectValidated
): UseVariableSuccess | UseVariableFailure => {
  const { id, pointer } = variable;
  const cacheKey = [id];
  const queryClient = useQueryClient();
  const queryState = queryClient.getQueryState(cacheKey);

  if (queryState === undefined) {
    return [undefined, "queryIdNotFound"];
  }

  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "couldNotStringify"];
  }

  return [dataParsed.data, undefined];
};

export const QueryVariable = ({ variable }: TemplateStringProps) => {
  const { t } = useTranslation();
  const mode = useMode();
  const [variableContent, error] = useQueryVariable(variable);

  if (error) {
    return (
      <Error value={variable.src}>
        {t(`direktivPage.error.templateString.query.${error}`)}
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
