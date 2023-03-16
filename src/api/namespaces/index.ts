import { NamespaceListSchema } from "./schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "../utils";
import { useApiKey } from "../../util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

export const getNamespaces = apiFactory({
  pathFn: () => `/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey: apiKey,
    params: undefined,
    pathParams: undefined,
  });

const namespaceKeys = {
  all: (apiKey: string) => [{ scope: "namespaces", apiKey }] as const,
};

export const useNamespaces = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: namespaceKeys.all(apiKey || "no-api-key"),
    queryFn: fetchNamespaces,
    staleTime: Infinity,
  });
};
