import { forceLeadingSlash, sortFoldersFirst } from "../utils";

import { NodeListSchema } from "../schema/node";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

// a node can be a directory or a file, the returned content could either
// be the list of files (if it's a direkctory) or the content of the file
const getNodeContent = apiFactory({
  url: ({
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
  schema: NodeListSchema,
});

const fetchTree = async ({
  queryKey: [{ apiKey, namespace, path, revision }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["nodeContent"]>>) =>
  getNodeContent({
    apiKey,
    urlParams: {
      namespace,
      path,
      revision,
    },
  });

export const useNodeContent = ({
  path,
  revision,
  enabled = true,
  namespace: givenNamespace,
}: {
  path?: string;
  revision?: string;
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
    enabled: !!namespace && enabled,
  });
};
