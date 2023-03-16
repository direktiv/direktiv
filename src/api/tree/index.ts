import type { QueryFunctionContext } from "@tanstack/react-query";
import { TreeListSchema } from "./schema";
import { apiFactory } from "../utils";
import { useApiKey } from "../../util/store/apiKey";
import { useNamespace } from "../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";

export const getNamespaces = apiFactory({
  pathFn: ({ namespace }: { namespace: string }) => `/api/${namespace}/tree`,
  method: "GET",
  schema: TreeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey: apiKey,
    params: undefined,
    pathParams: {
      namespace,
    },
  });

const namespaceKeys = {
  all: (apiKey: string, namespace: string) =>
    [{ scope: "tree", apiKey, namespace }] as const,
};

export const useTree = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  return useQuery({
    queryKey: namespaceKeys.all(
      apiKey || "no-api-key",
      namespace || "no-namespace"
    ),
    queryFn: fetchTree,
    networkMode: "always",
    enabled: !!namespace,
  });
};
