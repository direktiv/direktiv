import type { QueryFunctionContext } from "@tanstack/react-query";
import { VersionSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useQuery } from "@tanstack/react-query";
import { versionKeys } from "..";

export const getVersion = apiFactory({
  url: () => `/api/v2/version`,
  method: "GET",
  schema: VersionSchema,
});

const fetchVersions = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof versionKeys)["all"]>>) =>
  getVersion({
    apiKey,
    urlParams: undefined,
  });

export const useVersion = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: versionKeys.all(apiKey ?? undefined),
    queryFn: fetchVersions,
    staleTime: Infinity, // the api version shouldn't change
  });
};
