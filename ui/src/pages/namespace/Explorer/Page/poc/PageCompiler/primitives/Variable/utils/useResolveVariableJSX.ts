import {
  ResolveVariableFailure,
  useResolveVariable,
} from "./useResolveVariable";

import { VariableType } from "../../../../schema/primitives/variable";
import { z } from "zod";

export const JSXValueSchema = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
  z.undefined(),
]);

export type JSXValueType = z.infer<typeof JSXValueSchema>;

type ResolveVariableJSXSuccess = [JSXValueType, undefined];
type JSXFailure = [undefined, "couldNotStringify"];
type ResolveVariableJSXFailure = ResolveVariableFailure | JSXFailure;

export const useResolveVariableJSX = (
  value: VariableType
): ResolveVariableJSXSuccess | ResolveVariableJSXFailure => {
  const [data, error] = useResolveVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "couldNotStringify"];
  }

  return [dataParsed.data, undefined];
};
