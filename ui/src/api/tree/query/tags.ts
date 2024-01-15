import type { QueryFunctionContext } from "@tanstack/react-query";
import { TagsListSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getTags = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=tags`,
  method: "GET",
  schema: TagsListSchema,
});

const fetchRevisions = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["tagsList"]>>) =>
  getTags({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useNodeTags = ({
  path,
}: {
  path?: string;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.tagsList(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchRevisions,
    enabled: !!namespace,
  });
};
