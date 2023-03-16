import type { QueryFunctionContext } from "@tanstack/react-query";
import { TreeListSchema } from "./schema";
import { apiFactory } from "../utils";
import { useApiKey } from "../../util/store/apiKey";
import { useNamespace } from "../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { useToast } from "../../componentsNext/Toast";

const getTree = apiFactory({
  pathFn: ({
    namespace,
    directory,
  }: {
    namespace: string;
    directory?: string;
  }) => `/api/namespaces/${namespace}/tree${directory ? `/${directory}` : ""}`,
  method: "GET",
  schema: TreeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace, directory }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getTree({
    apiKey: apiKey,
    params: undefined,
    pathParams: {
      namespace,
      directory,
    },
  });

const namespaceKeys = {
  all: (apiKey: string, namespace: string, directory: string) =>
    [{ scope: "tree", apiKey, namespace, directory }] as const,
};

export const useTree = ({
  directory,
}: {
  directory?: string;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

  return useQuery({
    queryKey: namespaceKeys.all(
      apiKey ?? "no-api-key",
      namespace ?? "no-namespace",
      directory ?? ""
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
