import { InjectedVariables } from "../Variable/VariableContext";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";

// TODO: add useExtendedKeyValueArrayResolver
type ResolverFunction<DataType> = (
  value: DataType,
  options?: { variables: InjectedVariables }
) => DataType;

export const useKeyValueArrayResolver = (): ResolverFunction<
  KeyValueType[]
> => {
  const interpolateString = useStringInterpolation();
  return (input, options) =>
    input.map(({ key, value }) => {
      const parsedValue = interpolateString(value, options);
      return { key, value: parsedValue };
    });
};

const keyValueToObject = (kv: KeyValueType) => ({
  [kv.key]: kv.value,
});

export const keyValueArrayToObject = (kv: KeyValueType[]) =>
  kv.reduce((acc, curr) => ({ ...acc, ...keyValueToObject(curr) }), {});
