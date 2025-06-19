import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { useKeyValueArrayResolver } from "../primitives/keyValue/utils";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

/**
 * A hook that returns a function to generates a URL from a query or mutation
 */
export const useUrlGenerator = () => {
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const interpolateString = useStringInterpolation();

  return (input: QueryType | MutationType) => {
    const { url, queryParams } = input;
    const queryParamsResolved = resolveKeyValueArray(queryParams ?? []);
    const searchParams = new URLSearchParams();
    queryParamsResolved?.forEach(({ key, value }) => {
      searchParams.append(key, value);
    });

    const queryString = searchParams.toString();
    const interpolatedUrl = interpolateString(url);

    const requestUrl = queryString
      ? interpolatedUrl.concat("?", queryString)
      : interpolatedUrl;

    return requestUrl;
  };
};
