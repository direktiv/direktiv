import { ResolveVariableArrayError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

const UnknownArraySchema = z.array(z.unknown());
type UnknownArray = z.infer<typeof UnknownArraySchema>;

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is an array.
 */
export const useVariableArrayResolver = () => {
  const resolveVariable = useVariableResolver();
  return (
    value: VariableType
  ): Result<UnknownArray, ResolveVariableArrayError> => {
    const variableResult = resolveVariable(value);
    if (!variableResult.success) {
      return { success: false, error: variableResult.error };
    }

    const dataParsed = UnknownArraySchema.safeParse(variableResult.data);
    if (!dataParsed.success) {
      return { success: false, error: "notAnArray" };
    }

    return { success: true, data: dataParsed.data };
  };
};
