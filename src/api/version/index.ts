import type { QueryFunctionContext } from "@tanstack/react-query";
import { VersionSchema } from "./schema";
import { apiFactory } from "../utils";
import { useApiKeyState } from "../../util/store";
import { useQuery } from "@tanstack/react-query";

const getVersion = apiFactory({
  path: `/api/version`,
  method: "GET",
  schema: VersionSchema,
});

const fetchVersions = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof versionKeys)["all"]>>) =>
  getVersion({
    apiKey: apiKey,
    params: undefined,
  });

const versionKeys = {
  all: (apiKey: string) => [{ scope: "versions", apiKey }] as const,
};

export const useVersion = () => {
  const apiKey = useApiKeyState((state) => state.apiKey);
  return useQuery({
    queryKey: versionKeys.all(apiKey || "no-api-key"),
    queryFn: fetchVersions,
    networkMode: "always",
    staleTime: Infinity,
  });
};
