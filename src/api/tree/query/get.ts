import type { QueryFunctionContext } from "@tanstack/react-query";
import { TreeListSchema } from "../schema";
import { apiFactory } from "../../utils";
import { namespaceKeys } from "../";
import { sortFoldersFirst } from "../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { useToast } from "../../../componentsNext/Toast";

const getTree = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${path ? `/${path}` : ""}`,
  method: "GET",
  schema: TreeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getTree({
    apiKey: apiKey,
    params: undefined,
    pathParams: {
      namespace,
      path,
    },
  });

export const useListDirectoy = ({
  path,
}: {
  path?: string;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

  return useQuery({
    queryKey: namespaceKeys.all(
      apiKey ?? "no-api-key",
      namespace ?? "no-namespace",
      path ?? ""
    ),
    queryFn: fetchTree,
    select(data) {
      if (data?.children?.results) {
        return {
          ...data,
          children: {
            ...data.children,
            results: data.children.results.sort(sortFoldersFirst),
          },
        };
      }
      return data;
    },

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
