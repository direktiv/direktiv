import type { QueryFunctionContext } from "@tanstack/react-query";
import { RevisionsListSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getRevisions = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=refs`,
  method: "GET",
  schema: RevisionsListSchema,
});

const fetchRevisions = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["revisionsList"]>>) =>
  getRevisions({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useNodeRevisions = ({
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
    queryKey: treeKeys.revisionsList(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchRevisions,
    enabled: !!namespace,
  });
};
