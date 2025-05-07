import {
  GetValueFromJsonPathError,
  PossibleValues,
  ValidateVariableError,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";

import { VariableType } from "../../../../schema/primitives/variable";
import { useVariables } from "../VariableContext";

type ResolveVariableResult = [PossibleValues, undefined];

export type VariableError = [undefined, "NoStateForId"];

export type ResolveVariableError =
  | GetValueFromJsonPathError
  | ValidateVariableError
  | VariableError;

/**
 * Resolves a variable path string to its corresponding value stored in React context.
 *
 * Takes a variable string (e.g. "query.company-list.data.0.name") that specifies the
 * namespace, ID, and JSON pointer to retrieve the value.
 *
 * returns a tuple containing either [value, undefined] on success or [undefined, error]
 * on error, where error code describes the reason
 */
export const useResolveVariable = (
  value: VariableType
): ResolveVariableResult | ResolveVariableError => {
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
