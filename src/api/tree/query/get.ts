import { forceLeadingSlash, sortFoldersFirst } from "../utils";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { TreeListSchema } from "../schema";
import { apiFactory } from "../../utils";
import { treeKeys } from "../";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getTree = apiFactory({
  pathFn: ({
    namespace,
    path,
    revision,
  }: {
    namespace: string;
    path?: string;
    revision?: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}${
      revision ? `?ref=${revision}` : ""
    }`,
  method: "GET",
  schema: TreeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace, path, revision }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["nodeContent"]>>) =>
  getTree({
    apiKey: apiKey,
    params: undefined,
    pathParams: {
      namespace,
      path,
      revision,
    },
  });

export const useNodeContent = ({
  path,
  revision,
}: {
  path?: string;
  revision?: string;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.nodeContent(namespace, {
      apiKey: apiKey ?? undefined,
      path,
      revision,
    }),
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
  });
};
