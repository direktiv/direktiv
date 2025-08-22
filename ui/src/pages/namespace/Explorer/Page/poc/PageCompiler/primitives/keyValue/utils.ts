import { ExtendedKeyValueType } from "../../../schema/primitives/extendedKeyValue";
import { JsonValueType } from "../Variable/utils";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { LocalVariablesContent } from "../Variable/LocalVariables";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";
import { useTranslation } from "react-i18next";
import { useVariableResolver } from "../Variable/utils/useVariableResolver";

type KeyValueResolverFunction<InputType, OutputType = InputType> = (
  value: InputType,
  localVariables?: LocalVariablesContent
) => OutputType;

export const useKeyValueArrayResolver = (): KeyValueResolverFunction<
  KeyValueType[]
> => {
  const interpolateString = useStringInterpolation();
  return (input, localVariables) =>
    input.map(({ key, value }) => {
      const parsedValue = interpolateString(value, localVariables);
      return { key, value: parsedValue };
    });
};

const keyValueToObject = (kv: KeyValueType) => ({
  [kv.key]: kv.value,
});

export const keyValueArrayToObject = (kv: KeyValueType[]) =>
  kv.reduce((acc, curr) => ({ ...acc, ...keyValueToObject(curr) }), {});

type ExtendedKeyValueCompiledType = {
  key: string;
  value: JsonValueType;
};

export const useExtenedKeyValueArrayResolver = (): KeyValueResolverFunction<
  ExtendedKeyValueType[],
  ExtendedKeyValueCompiledType[]
> => {
  const { t } = useTranslation();
  const interpolateString = useStringInterpolation();
  const resolveVariable = useVariableResolver();
  return (input, localVariables) =>
    input.map(({ key, value }) => {
      switch (value.type) {
        case "string": {
          return { key, value: interpolateString(value.value, localVariables) };
        }
        case "variable": {
          const resolvedVariable = resolveVariable(value.value, localVariables);
          if (!resolvedVariable.success) {
            throw new Error(
              t(`direktivPage.error.templateString.${resolvedVariable.error}`, {
                variable: value.value,
              })
            );
          }
          return { key, value: resolvedVariable.data };
        }
        case "object": {
          return { key, value: keyValueArrayToObject(value.value) };
        }
        case "boolean":
        case "string-array":
        case "boolean-array":
        case "number-array":
        case "number":
          {
            return { key, value: value.value };
          }

          throw new Error(
            `${value.type} is not implemented for extended key value`
          );
      }
    });
};
