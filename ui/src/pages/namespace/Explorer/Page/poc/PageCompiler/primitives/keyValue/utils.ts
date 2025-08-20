import { KeyValueType } from "../../../schema/primitives/keyValue";
import { LocalVariables } from "../Variable/LocalVariables";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";

// TODO: add useExtendedKeyValueArrayResolver
type ResolverFunction<DataType> = (
  value: DataType,
  localVariables?: LocalVariables
) => DataType;

export const useKeyValueArrayResolver = (): ResolverFunction<
  KeyValueType[]
> => {
  const interpolateString = useStringInterpolation();
  return (input, formVariables) =>
    input.map(({ key, value }) => {
      const parsedValue = interpolateString(value, formVariables);
      return { key, value: parsedValue };
    });
};

const keyValueToObject = (kv: KeyValueType) => ({
  [kv.key]: kv.value,
});

export const keyValueArrayToObject = (kv: KeyValueType[]) =>
  kv.reduce((acc, curr) => ({ ...acc, ...keyValueToObject(curr) }), {});
