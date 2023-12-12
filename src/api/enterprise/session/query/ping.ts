import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { SessionSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import enterpriseConfig from "~/config/enterprise";
import { getPermissionStatus } from "~/api/errorHandling";
import { sessionKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";

const getSession = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/ping`,
  method: "GET",
  schema: SessionSchema,
});

const fetchSession = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof sessionKeys)["get"]>>) =>
  getSession({
    apiKey,
    urlParams: {},
  });

export const useRefreshSession = ({ enabled }: { enabled: boolean }) => {
  const apiKey = useApiKey();

  const { isError, error } = useQuery({
    queryKey: sessionKeys.get(apiKey ?? undefined),
    queryFn: fetchSession,
    enabled,
    refetchInterval: 1000 * 10,
    refetchIntervalInBackground: true,
  });

  if (isError && getPermissionStatus(error).isAllowed === false) {
    window.location.href = enterpriseConfig.logoutPath;
  }
};
