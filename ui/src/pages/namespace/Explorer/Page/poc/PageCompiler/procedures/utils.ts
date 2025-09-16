import { LocalVariablesContent } from "../primitives/Variable/VariableContext";
import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { keyValueArrayToObject } from "../primitives/keyValue/utils";
import { useKeyValueArrayResolver } from "../primitives/keyValue/useKeyValueArrayResolver";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

/**
 * A hook that returns a function to generate a URL from a query or mutation
 */
export const useUrlGenerator = () => {
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const interpolateString = useStringInterpolation();

  return (
    input: QueryType | MutationType,
    variables?: LocalVariablesContent
  ) => {
    const { url, queryParams } = input;
    const queryParamsResolved = resolveKeyValueArray(
      queryParams ?? [],
      variables
    );
    const searchParams = new URLSearchParams(
      keyValueArrayToObject(queryParamsResolved)
    );
    const paramsString = searchParams.toString();

    const interpolatedUrl = interpolateString(url, variables);

    const requestUrl = paramsString
      ? interpolatedUrl.concat("?", paramsString)
      : interpolatedUrl;

    return requestUrl;
  };
};
