import { ResolveVariableArrayError } from "./errors";
import { ResolverFunction } from "./types";
import { useVariableResolver } from "./useVariableResolver";
import { z } from "zod";

const UnknownArraySchema = z.array(z.unknown());
type UnknownArray = z.infer<typeof UnknownArraySchema>;

/**
 * A hook that works the same as useVariableResolver
 * but ensures that the resolved value is an array.
 */
export const useVariableArrayResolver = (): ResolverFunction<
  UnknownArray,
  ResolveVariableArrayError
> => {
  const resolveVariable = useVariableResolver();
  return (...args) => {
    const variableResult = resolveVariable(...args);
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
