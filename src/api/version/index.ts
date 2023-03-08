import type { QueryFunctionContext } from "@tanstack/react-query";
import { VersionSchema } from "./schema";
import { apiFactory } from "../utils";
import { useQuery } from "@tanstack/react-query";

const getVersion = apiFactory({
  path: `/api/version`,
  method: "GET",
  schema: VersionSchema,
});

const fetchVersions = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof versionKeys)["list"]>>) =>
  getVersion({
    apiKey: apiKey,
    params: undefined,
  });

const versionKeys = {
  all: [{ scope: "versions" }] as const,
  list: (apiKey: string) => [{ ...versionKeys.all[0], apiKey }] as const,
};

export const useVersion = () =>
  useQuery({
    queryKey: versionKeys.list("password"),
    queryFn: fetchVersions,
    networkMode: "always",
    staleTime: Infinity,
  });
