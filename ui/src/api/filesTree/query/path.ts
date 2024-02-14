import { PathListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
import { nodeKeys } from "..";
import { sortFoldersFirst } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getPath = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/v2/namespaces/${namespace}/files-tree${forceLeadingSlash(path)}`,
  method: "GET",
  schema: PathListSchema,
});

const fetchPath = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof nodeKeys)["nodesList"]>>) =>
  getPath({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const usePath = ({
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
    queryKey: nodeKeys.nodesList(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchPath,
    select(data) {
      return data.data.paths.sort(sortFoldersFirst);
    },
    enabled: !!namespace && enabled,
  });
};
