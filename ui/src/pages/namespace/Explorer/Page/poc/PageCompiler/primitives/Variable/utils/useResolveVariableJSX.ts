import { ResolveVariableError, useResolveVariable } from "./useResolveVariable";

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

type ResolveVariableJSXResult = [JSXValueType, undefined];
type JSXError = [undefined, "couldNotStringify"];
type ResolveVariableJSXError = ResolveVariableError | JSXError;

export const useResolveVariableJSX = (
  value: VariableType
): ResolveVariableJSXResult | ResolveVariableJSXError => {
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
