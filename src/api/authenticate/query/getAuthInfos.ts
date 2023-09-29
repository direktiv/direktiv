import { QueryFunctionContext, useQuery } from "@tanstack/react-query";
import { authenticationKeys, checkApiKeyAgainstServer } from "..";

const authTest = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<
  ReturnType<(typeof authenticationKeys)["authentication"]>
>) => checkApiKeyAgainstServer(apiKey);

export const useAuthTest = ({
  apikey,
  enabled,
}: { apikey?: string; enabled?: boolean } = {}) =>
  useQuery({
    queryKey: authenticationKeys.authentication(apikey),
    queryFn: authTest,
    enabled,
  });
