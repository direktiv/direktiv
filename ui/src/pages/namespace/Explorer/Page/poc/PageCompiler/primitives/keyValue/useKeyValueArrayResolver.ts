import { KeyValueResolverFunction } from "./utils";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";

export const useKeyValueArrayResolver = (): KeyValueResolverFunction<
  KeyValueType[]
> => {
  const interpolateString = useStringInterpolation();
  return (keyValueArray, localVariables) =>
    keyValueArray.map(({ key, value }) => {
      const parsedValue = interpolateString(value, localVariables);
      return { key, value: parsedValue };
    });
};
