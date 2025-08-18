import { ResolveVariableStringError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
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
export const useVariableStringResolver = () => {
  const resolveVariable = useVariableResolver();
  return (value: VariableType): Result<string, ResolveVariableStringError> => {
    const variableResult = resolveVariable(value);

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
