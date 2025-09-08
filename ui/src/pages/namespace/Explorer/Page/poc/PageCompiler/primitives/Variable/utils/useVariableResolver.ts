import { JsonPathError, ResolveVariableError } from "./errors";
import {
  JsonValueType,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";
import { ResolverFunction, ValidationResult } from "./types";

import { localVariableNamespace } from "../../../../schema/primitives/variable";
import { useVariablesContext } from "../VariableContext";

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
  const contextVariables = useVariablesContext();

  return (value, localVariables) => {
    const variableObject = parseVariable(value);
    const validationResult = validateVariable(variableObject);
    if (!validationResult.success) {
      return { success: false, error: validationResult.error };
    }
    const { id, pointer, namespace } = validationResult.data;

    // let variableValue: unknown;
    let jsonPathResult: ValidationResult<JsonValueType, JsonPathError>;

    if (namespace === localVariableNamespace) {
      if (localVariables === undefined) {
        throw new Error(
          "You can't access variables from the 'this' namespace in this context"
        );
      }

      jsonPathResult = localVariables[id];

      return { success: true, data: jsonPathResult };
    } else {
      if (!contextVariables[namespace][id]) {
        return { success: false, error: "NoStateForId" };
      }
      jsonPathResult = getValueFromJsonPath(
        contextVariables[namespace][id],
        pointer
      );
    }

    if (!jsonPathResult.success) {
      return { success: false, error: jsonPathResult.error };
    }

    return { success: true, data: jsonPathResult.data };
  };
};
