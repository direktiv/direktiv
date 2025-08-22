import { KeyValueResolverFunction, keyValueArrayToObject } from "./utils";

import { ExtendedKeyValueType } from "../../../schema/primitives/extendedKeyValue";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";
import { useTranslation } from "react-i18next";
import { useVariableResolver } from "../Variable/utils/useVariableResolver";

export const useExtenedKeyValueArrayResolver = (): KeyValueResolverFunction<
  ExtendedKeyValueType[]
> => {
  const { t } = useTranslation();
  const interpolateString = useStringInterpolation();
  const resolveVariable = useVariableResolver();
  return (keyValueArray, localVariables) =>
    keyValueArray.map(({ key, value: valueType }) => {
      switch (valueType.type) {
        case "string": {
          return {
            key,
            value: interpolateString(valueType.value, localVariables),
          };
        }
        case "variable": {
          const resolvedVariable = resolveVariable(
            valueType.value,
            localVariables
          );
          if (!resolvedVariable.success) {
            throw new Error(
              t(`direktivPage.error.templateString.${resolvedVariable.error}`, {
                variable: valueType.value,
              })
            );
          }
          return { key, value: resolvedVariable.data };
        }
        case "object": {
          return { key, value: keyValueArrayToObject(valueType.value) };
        }
        case "boolean":
        case "string-array":
        case "boolean-array":
        case "number-array":
        case "number":
          {
            return { key, value: valueType.value };
          }

          throw new Error(
            `${valueType.type} is not implemented for extended key value`
          );
      }
    });
};
