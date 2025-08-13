import { ResolveVariableNumberError } from "./errors";
import { ResolverFunctionWithError } from "./types";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is a number.
 */
export const useVariableNumberResolver = (): ResolverFunctionWithError<
  number,
  ResolveVariableNumberError
> => {
  const resolveVariable = useVariableResolver();
  return (...args) => {
    const variableResult = resolveVariable(...args);

    if (!variableResult.success) {
      return { success: false, error: variableResult.error };
    }

    const dataParsed = z.number().safeParse(variableResult.data);
    if (!dataParsed.success) {
      return { success: false, error: "notANumber" };
    }

    return { success: true, data: dataParsed.data };
  };
};
