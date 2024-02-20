import { FileListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
import { pathKeys } from "..";
import { sortFoldersFirst } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getFile = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/v2/namespaces/${namespace}/files${forceLeadingSlash(path)}`,
  method: "GET",
  schema: FileListSchema,
});

const fetchFile = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof pathKeys)["paths"]>>) =>
  getFile({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useFile = ({
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
    queryKey: pathKeys.paths(namespace, {
      apiKey: apiKey ?? undefined,
      path: forceLeadingSlash(path),
    }),
    queryFn: fetchFile,
    select(response) {
      const { data } = response;
      if (data.type === "directory") {
        data.children = data.children.sort(sortFoldersFirst);
      }
      return data;
    },
    enabled: !!namespace && enabled,
  });
};
