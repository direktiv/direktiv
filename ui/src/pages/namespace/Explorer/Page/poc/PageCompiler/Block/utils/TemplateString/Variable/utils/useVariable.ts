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
import { useVariables } from "../../../../../store/variables";

type UseVariableSuccess = [PossibleValues, undefined];
export type VariableFailure = [undefined, "queryIdNotFound" | "loopIdNotFound"];

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

  const variables = useVariables();

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
      const [queryData, queryError] = getValueFromJsonPath(cachedData, pointer);

      if (queryError) {
        return [undefined, queryError];
      }

      return [queryData, undefined];
    }

    case "loop": {
      if (!variables["loop"][id]) {
        return [undefined, "loopIdNotFound"];
      }

      const [loopData, loopError] = getValueFromJsonPath(
        variables["loop"][id],
        // TODO: make index dynamic here, by using a react context?
        `0.${pointer}`
      );

      if (loopError) {
        return [undefined, loopError];
      }

      return [loopData, undefined];
    }
  }
};
