import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { authenticationKeys } from "..";
import { getVersion } from "~/api/version/query/get";

const authTest = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<
  ReturnType<(typeof authenticationKeys)["authentication"]>
>) =>
  getVersion({
    apiKey,
    urlParams: undefined,
  }) // when test call succeeds, and matches the schema out auth test passes
    .then(() => true)
    // when the test call fails with a 401,
    .catch((err) => {
      if (err !== 401) {
        // any other error should bubble up, like a 500 or schema validation error
        throw err;
      }
      return false;
    });

export const useAuthTest = ({
  apikey,
  enabled,
}: { apikey?: string; enabled?: boolean } = {}) =>
  useQuery({
    queryKey: authenticationKeys.authentication(apikey),
    queryFn: authTest,
    enabled,
    // TODO: use stale time to infinity?
    // TODO: handle errors
  });
