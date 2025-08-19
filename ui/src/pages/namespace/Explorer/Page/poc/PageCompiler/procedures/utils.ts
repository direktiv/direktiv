import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";

import { InjectedVariables } from "../primitives/Variable/VariableContext";
import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

/**
 * A hook that returns a function to generate a URL from a query or mutation
 */
export const useUrlGenerator = () => {
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const interpolateString = useStringInterpolation();

  return (
    input: QueryType | MutationType,
    options?: { variables: InjectedVariables }
  ) => {
    const { url, queryParams } = input;
    const queryParamsResolved = resolveKeyValueArray(
      queryParams ?? [],
      options
    );
    const searchParams = new URLSearchParams(
      keyValueArrayToObject(queryParamsResolved)
    );
    const paramsString = searchParams.toString();

    const interpolatedUrl = interpolateString(url, options);

    const requestUrl = paramsString
      ? interpolatedUrl.concat("?", paramsString)
      : interpolatedUrl;

    return requestUrl;
  };
};
