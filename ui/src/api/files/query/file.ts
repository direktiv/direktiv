import { forceLeadingSlash, sortFoldersFirst } from "../utils";

import { FileListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { fileKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getFile = apiFactory({
  url: ({
    namespace,
    path,
    baseUrl,
  }: {
    namespace: string;
    path?: string;
    baseUrl?: string;
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/files${forceLeadingSlash(
      path
    )}`,
  method: "GET",
  schema: FileListSchema,
});

const fetchFile = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof fileKeys)["file"]>>) =>
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

  return useQueryWithPermissions({
    queryKey: fileKeys.file(namespace ?? "", {
      apiKey: apiKey ?? undefined,
      path: forceLeadingSlash(path),
    }),
    queryFn: fetchFile,
    select(response) {
      const { data } = response;
      if (data.type === "directory" && data.children) {
        data.children = data.children.sort(sortFoldersFirst);
      }
      return data;
    },
    enabled: !!namespace && enabled,
  });
};
