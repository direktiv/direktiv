import { UseVariableFailure, useVariable } from "./useVariable";

import { VariableType } from "../../../../../../schema/primitives/variable";
import { z } from "zod";

export const ArraySchema = z.array(z.unknown());

export type ArrayType = z.infer<typeof ArraySchema>;

type UseVariableArraySuccess = [ArrayType, undefined];
type ArrayFailure = [undefined, "notAnArray"];
type UseVariableArrayFailure = UseVariableFailure | ArrayFailure;

/**
 * useVariableArray does the same as useVariable, but it will add a validation
 * layer on top to ensure that the value returned is an array.
 */
export const useVariableArray = (
  value: VariableType
): UseVariableArraySuccess | UseVariableArrayFailure => {
  const [data, error] = useVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = ArraySchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "notAnArray"];
  }

  return [dataParsed.data, undefined];
};
