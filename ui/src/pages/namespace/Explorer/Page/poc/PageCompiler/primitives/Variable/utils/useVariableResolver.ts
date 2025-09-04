import {
  JsonValueType,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";

import { ResolveVariableError } from "./errors";
import { ResolverFunction } from "./types";
import { useMergeLocalWithContextVariables } from "../LocalVariables";

/**
 * A hook that returns a function to resolve a variable path string to its
 * corresponding value stored in React context.
 *
 * Takes a variable string (e.g. "query.company-list.data.0.name") that specifies
 * the namespace, ID, and JSON pointer to retrieve the value.
 *
 * Returns a Result object that indicates either success with the resolved value
 * or failure with an error code describing the reason
 */
export const useVariableResolver = (): ResolverFunction<
  JsonValueType,
  ResolveVariableError
> => {
  const mergeLocalWithContextVariables = useMergeLocalWithContextVariables();
  return (value, localVariables) => {
    const variables = mergeLocalWithContextVariables(localVariables);
    const variableObject = parseVariable(value);
    const validationResult = validateVariable(variableObject);

    if (!validationResult.success) {
      return { success: false, error: validationResult.error };
    }

    const { id, pointer, namespace } = validationResult.data;

    if (!variables[namespace]?.[id]) {
      return { success: false, error: "NoStateForId" };
    }

    const jsonPathResult = getValueFromJsonPath(
      variables[namespace]?.[id],
      pointer
    );

    if (!jsonPathResult.success) {
      return { success: false, error: jsonPathResult.error };
    }

    return { success: true, data: jsonPathResult.data };
  };
};
