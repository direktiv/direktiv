import {
  GetValueFromJsonPathFailure,
  PossibleValues,
  ValidateVariableFailure,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";

import { VariableType } from "../../../../schema/primitives/variable";
import { useVariables } from "../VariableContext";

type ResolveVariableSuccess = [PossibleValues, undefined];

export type VariableFailure = [undefined, "NoStateForId"];

export type ResolveVariableFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | VariableFailure;

/**
 * Resolves a variable path string to its corresponding value stored in React context.
 *
 * Takes a variable string (e.g. "query.company-list.data.0.name") that specifies the
 * namespace, ID, and JSON pointer to retrieve the value.
 *
 * returns a tuple containing either [value, undefined] on success or [undefined, error]
 * on failure, where error code describes the failure reason
 */
export const useResolveVariable = (
  value: VariableType
): ResolveVariableSuccess | ResolveVariableFailure => {
  const [variableObject, validationError] = validateVariable(
    parseVariable(value)
  );
  const variables = useVariables();

  if (validationError) {
    return [undefined, validationError];
  }

  const { id, pointer, namespace } = variableObject;

  if (!variables[namespace][id]) {
    return [undefined, "NoStateForId"];
  }

  const [data, error] = getValueFromJsonPath(variables[namespace][id], pointer);

  if (error) {
    return [undefined, error];
  }

  return [data, undefined];
};
