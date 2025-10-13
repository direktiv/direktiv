import { ResolveVariableStringArrayError } from "./errors";
import { ResolverFunction } from "./types";
import { useVariableArrayResolver } from "./useVariableArrayResolver";
import { z } from "zod";

const StringArraySchema = z.array(z.string());
type UnknownArray = z.infer<typeof StringArraySchema>;

/**
 * A hook that works the same as useVariableArrayResolver
 * but ensures that the resolved array contains only strings.
 */
export const useVariableStringArrayResolver = (): ResolverFunction<
  UnknownArray,
  ResolveVariableStringArrayError
> => {
  const resolveVariable = useVariableArrayResolver();
  return (...args) => {
    const variableResult = resolveVariable(...args);
    if (!variableResult.success) {
      return { success: false, error: variableResult.error };
    }

    const dataParsed = StringArraySchema.safeParse(variableResult.data);
    if (!dataParsed.success) {
      return { success: false, error: "notAnArrayOfStrings" };
    }

    return { success: true, data: dataParsed.data };
  };
};
