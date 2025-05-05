import {
  GetValueFromJsonPathFailure,
  PossibleValues,
  ValidateVariableFailure,
  getValueFromJsonPath,
  parseVariable,
  validateVariable,
} from ".";

import { VariableType } from "../../../../../../schema/primitives/variable";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

type UseVariableSuccess = [PossibleValues, undefined];
type QueryFailure = [undefined, "queryIdNotFound" | "couldNotStringify"];

type UseVariableFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | QueryFailure;

/**
 * useVariable takes a variable string like "query.company-list.data.0.name" and
 * returns the value at the specified path.
 */
export const useVariable = (
  value: VariableType
): UseVariableSuccess | UseVariableFailure => {
  const queryClient = useQueryClient();
  const [variableObject, validationError] = validateVariable(
    parseVariable(value)
  );

  if (validationError) {
    return [undefined, validationError];
  }

  const { id, pointer } = variableObject;
  const cacheKey = [id];
  const queryState = queryClient.getQueryState(cacheKey);

  if (queryState === undefined) {
    return [undefined, "queryIdNotFound"];
  }

  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);

  if (error) {
    return [undefined, error];
  }

  return [data, undefined];
};

export const JSXValueSchema = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
  z.undefined(),
]);

export type JSXValueType = z.infer<typeof JSXValueSchema>;

type UseVariableJSXSuccess = [JSXValueType, undefined];

type UseVariableJSXFailure =
  | GetValueFromJsonPathFailure
  | ValidateVariableFailure
  | QueryFailure;

/**
 * useVariableJSX does the same as useVariable, but it will add some validation on top to ensure
 * that the value returned is JSX compatible.
 */
export const useVariableJSX = (
  value: VariableType
): UseVariableJSXSuccess | UseVariableJSXFailure => {
  const [data, error] = useVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "couldNotStringify"];
  }

  return [dataParsed.data, undefined];
};
