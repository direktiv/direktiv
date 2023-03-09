import { NamespaceListSchema } from "./schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "../utils";
import { useApiKeyState } from "../../util/store";
import { useQuery } from "@tanstack/react-query";

export const getNamespaces = apiFactory({
  path: `/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey: apiKey,
    params: undefined,
  });

const namespaceKeys = {
  all: (apiKey: string) => [{ scope: "namespaces", apiKey }] as const,
};

export const useNamespaces = () => {
  const apiKey = useApiKeyState((state) => state.apiKey);
  return useQuery({
    queryKey: namespaceKeys.all(apiKey || "no-api-key"),
    queryFn: fetchNamespaces,
    networkMode: "always",
    staleTime: Infinity,
  });
};
