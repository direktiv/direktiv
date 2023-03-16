import type { QueryFunctionContext } from "@tanstack/react-query";
import { TreeListSchema } from "./schema";
import { apiFactory } from "../utils";
import { useApiKey } from "../../util/store/apiKey";
import { useNamespace } from "../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { useToast } from "../../componentsNext/Toast";

const getTree = apiFactory({
  pathFn: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/tree`,
  method: "GET",
  schema: TreeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getTree({
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
  const { toast } = useToast();

  return useQuery({
    queryKey: namespaceKeys.all(
      apiKey || "no-api-key",
      namespace || "no-namespace"
    ),
    queryFn: fetchTree,
    enabled: !!namespace,
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not fetch tree 😢",
        variant: "error",
      });
    },
  });
};
