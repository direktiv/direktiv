import { apiFactory, defaultKeys } from "../../utils";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { RevisionsListSchema } from "../schema";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "../";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { useToast } from "../../../design/Toast";

const getRevisions = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=refs`,
  method: "GET",
  schema: RevisionsListSchema,
});

const fetchRevisions = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<ReturnType<(typeof treeKeys)["revisionsList"]>>) =>
  getRevisions({
    apiKey: apiKey,
    params: undefined,
    pathParams: {
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
  const { toast } = useToast();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.revisionsList(
      apiKey ?? defaultKeys.apiKey,
      namespace,
      path ?? ""
    ),
    queryFn: fetchRevisions,
    enabled: !!namespace,
    // TODO: remove to global error handler
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not fetch directory content 😢",
        variant: "error",
      });
    },
  });
};
