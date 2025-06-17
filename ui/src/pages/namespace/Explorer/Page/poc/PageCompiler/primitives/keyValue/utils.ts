import { KeyValueType } from "../../../schema/primitives/keyValue";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";

export const useKeyValueArrayResolver = () => {
  const interpolateString = useStringInterpolation();
  return (input: KeyValueType[]): KeyValueType[] =>
    input.map(({ key, value }) => {
      const parsedValue = interpolateString(value);
      return { key, value: parsedValue };
    });
};
