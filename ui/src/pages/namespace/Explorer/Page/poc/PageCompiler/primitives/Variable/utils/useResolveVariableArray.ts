import {
  ResolveVariableFailure,
  useResolveVariable,
} from "./useResolveVariable";

import { VariableType } from "../../../../schema/primitives/variable";
import { z } from "zod";

export const ArraySchema = z.array(z.unknown());

export type ArrayType = z.infer<typeof ArraySchema>;

type ResolveVariableArraySuccess = [ArrayType, undefined];
type ArrayFailure = [undefined, "notAnArray"];
type ResolveVariableArrayFailure = ResolveVariableFailure | ArrayFailure;

export const useResolveVariableArray = (
  value: VariableType
): ResolveVariableArraySuccess | ResolveVariableArrayFailure => {
  const [data, error] = useResolveVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = ArraySchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "notAnArray"];
  }

  return [dataParsed.data, undefined];
};
