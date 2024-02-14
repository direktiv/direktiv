import { forceLeadingSlash, sortFoldersFirst } from "../utils";

import { NodeListSchema } from "../schema/node";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

// a node can be a directory or a file, the returned content could either
// be the list of files (if it's a directory) or the content of the file
const getNodeContent = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}`,
  method: "GET",
  schema: NodeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["nodeContent"]>>) =>
  getNodeContent({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useNodeContent = ({
  path,
  enabled = true,
  namespace: givenNamespace,
}: {
  path?: string;
  enabled?: boolean;
  namespace?: string;
} = {}) => {
  const defaultNamespace = useNamespace();

  const namespace = givenNamespace ? givenNamespace : defaultNamespace;
  const apiKey = useApiKey();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: treeKeys.nodeContent(namespace, {
      apiKey: apiKey ?? undefined,
      path,
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
    enabled: !!namespace && enabled,
  });
};
