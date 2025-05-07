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

// TODO: refine comment
/**
 * useResolveVariable takes a variable string like "query.company-list.data.0.name"
 * and returns the value at the specified path that is stored in react context.
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
