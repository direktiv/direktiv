import { KeyValueType } from "../../schema/primitives/keyValue";
import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

export const useGetUrl = () => {
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const interpolateString = useStringInterpolation();

  return (input: QueryType | MutationType) => {
    const { baseUrl, queryParams } = input;
    const queryParamsParsed = resolveKeyValueArray(queryParams ?? []);
    const searchParams = new URLSearchParams();
    queryParamsParsed?.forEach(({ key, value }) => {
      searchParams.append(key, value);
    });

    const queryString = searchParams.toString();
    const url = interpolateString(baseUrl);
    return queryString ? url.concat("?", queryString) : url;
  };
};

const useKeyValueArrayResolver = () => {
  const interpolateString = useStringInterpolation();
  return (input: KeyValueType[]): KeyValueType[] =>
    input.map(({ key, value }) => {
      const parsedValue = interpolateString(value);
      return { key, value: parsedValue };
    });
};
