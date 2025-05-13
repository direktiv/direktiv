import { ResolveVariableStringError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
import { useResolveVariable } from "./useResolveVariable";
import { z } from "zod";

export const StringCompatible = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
]);

export type StringCompatibleType = z.infer<typeof StringCompatible>;

export const useResolveVariableString = (
  value: VariableType
): Result<StringCompatibleType, ResolveVariableStringError> => {
  const variableResult = useResolveVariable(value);

  if (!variableResult.success) {
    return { success: false, error: variableResult.error };
  }

  const dataParsed = StringCompatible.safeParse(variableResult.data);
  if (!dataParsed.success) {
    return { success: false, error: "couldNotStringify" };
  }

  return { success: true, data: `${dataParsed.data}` };
};
