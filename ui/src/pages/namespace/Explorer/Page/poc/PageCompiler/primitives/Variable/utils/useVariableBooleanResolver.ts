import { ResolveVariableBooleanError } from "./errors";
import { ResolverFunctionWithError } from "./types";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is a boolean.
 */
export const useVariableBooleanResolver = (): ResolverFunctionWithError<
  boolean,
  ResolveVariableBooleanError
> => {
  const resolveVariable = useVariableResolver();
  return (...args) => {
    const variableResult = resolveVariable(...args);

    if (!variableResult.success) {
      return { success: false, error: variableResult.error };
    }

    const dataParsed = z.boolean().safeParse(variableResult.data);
    if (!dataParsed.success) {
      return { success: false, error: "notABoolean" };
    }

    return { success: true, data: dataParsed.data };
  };
};
