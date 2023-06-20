import type { QueryFunctionContext } from "@tanstack/react-query";
import { VersionSchema } from "./schema";
import { apiFactory } from "../apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

const getVersion = apiFactory({
  url: () => `/api/version`,
  method: "GET",
  schema: VersionSchema,
});

const fetchVersions = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof versionKeys)["all"]>>) =>
  getVersion({
    apiKey,
    payload: undefined,
    headers: undefined,
    urlParams: undefined,
  });

const versionKeys = {
  all: (apiKey: string | undefined) => [{ scope: "versions", apiKey }] as const,
};

export const useVersion = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: versionKeys.all(apiKey ?? undefined),
    queryFn: fetchVersions,
    staleTime: Infinity, // the api version shouldn't change
  });
};
