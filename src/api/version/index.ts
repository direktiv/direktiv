import type { QueryFunctionContext } from "@tanstack/react-query";
import { VersionSchema } from "./schema";
import { apiFactory } from "../utils";
import { useApiKey } from "../../util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

const getVersion = apiFactory({
  pathFn: () => `/api/version`,
  method: "GET",
  schema: VersionSchema,
});

const fetchVersions = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof versionKeys)["all"]>>) =>
  getVersion({
    apiKey: apiKey,
    params: undefined,
    pathParams: undefined,
  });

const versionKeys = {
  all: (apiKey: string) => [{ scope: "versions", apiKey }] as const,
};

export const useVersion = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: versionKeys.all(apiKey ?? "no-api-key"),
    queryFn: fetchVersions,
    staleTime: Infinity,
    networkMode: "always", // the default networkMode sometimes assumes that the client is offline
  });
};
