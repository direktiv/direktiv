import { ResolveVariableStringError } from "./errors";
import { ResolverFunctionWithError } from "./types";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

const StringCompatible = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
]);

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is a string.
 */
export const useVariableStringResolver = (): ResolverFunctionWithError<
  string,
  ResolveVariableStringError
> => {
  const resolveVariable = useVariableResolver();
  return (...args) => {
    const variableResult = resolveVariable(...args);

    if (!variableResult.success) {
      return { success: false, error: variableResult.error };
    }

    const dataParsed = StringCompatible.safeParse(variableResult.data);
    if (!dataParsed.success) {
      return { success: false, error: "couldNotStringify" };
    }

    return { success: true, data: String(dataParsed.data) };
  };
};
