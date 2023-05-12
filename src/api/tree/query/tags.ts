import type { QueryFunctionContext } from "@tanstack/react-query";
import { TagsListSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "../";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getTags = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=tags`,
  method: "GET",
  schema: TagsListSchema,
});

const fetchRevisions = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["tagsList"]>>) =>
  getTags({
    apiKey: apiKey,
    params: undefined,
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
