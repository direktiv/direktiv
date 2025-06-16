import { ResolveVariableArrayError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
import { useResolveVariable } from "./useResolveVariable";
import { z } from "zod";

const UnknownArraySchema = z.array(z.unknown());
type UnknownArray = z.infer<typeof UnknownArraySchema>;

export const useResolveVariableArray = () => {
  const variableResultFn = useResolveVariable();
  return (
    value: VariableType
  ): Result<UnknownArray, ResolveVariableArrayError> => {
    const variableResult = variableResultFn(value);
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
