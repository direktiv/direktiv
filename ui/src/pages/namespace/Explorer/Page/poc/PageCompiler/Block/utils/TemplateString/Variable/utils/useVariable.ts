import {
  GetValueFromJsonPathFailure,
  PossibleValues,
  ValidateVariableFailure,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";

import { VariableType } from "../../../../../../schema/primitives/variable";
import { useQueryClient } from "@tanstack/react-query";

type UseVariableSuccess = [PossibleValues, undefined];
export type VariableFailure = [undefined, "queryIdNotFound"];

export type UseVariableFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | VariableFailure;

/**
 * useVariable takes a variable string like "query.company-list.data.0.name" and
 * returns the value at the specified path.
 */
export const useVariable = (
  value: VariableType
): UseVariableSuccess | UseVariableFailure => {
  const queryClient = useQueryClient();
  const [variableObject, validationError] = validateVariable(
    parseVariable(value)
  );

  if (validationError) {
    return [undefined, validationError];
  }

  const { id, pointer, namespace } = variableObject;

  switch (namespace) {
    case "query": {
      const cacheKey = [id];
      const queryState = queryClient.getQueryState(cacheKey);

      if (queryState === undefined) {
        return [undefined, "queryIdNotFound"];
      }

      const cachedData = queryClient.getQueryData(cacheKey);
      const [data, error] = getValueFromJsonPath(cachedData, pointer);

      if (error) {
        return [undefined, error];
      }

      return [data, undefined];
    }
  }
};
