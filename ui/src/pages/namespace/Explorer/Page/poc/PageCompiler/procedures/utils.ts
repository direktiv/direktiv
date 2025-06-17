import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { useKeyValueArrayResolver } from "../primitives/keyValue/utils";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

export const useUrlGenerator = () => {
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const interpolateString = useStringInterpolation();

  return (input: QueryType | MutationType) => {
    const { baseUrl, queryParams } = input;
    const queryParamsResolved = resolveKeyValueArray(queryParams ?? []);
    const searchParams = new URLSearchParams();
    queryParamsResolved?.forEach(({ key, value }) => {
      searchParams.append(key, value);
    });

    const queryString = searchParams.toString();
    const interpolatedBaseUrl = interpolateString(baseUrl);

    return queryString
      ? interpolatedBaseUrl.concat("?", queryString)
      : interpolatedBaseUrl;
  };
};
