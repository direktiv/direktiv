import { ResolveVariableNumberError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is a number.
 */
export const useVariableNumberResolver = () => {
  const resolveVariable = useVariableResolver();
  return (value: VariableType): Result<number, ResolveVariableNumberError> => {
    const variableResult = resolveVariable(value);

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
