import { useMutation, useQueryClient } from "@tanstack/react-query";

import { TreeNodeDeletedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const deleteRevision = apiFactory({
  pathFn: ({
    namespace,
    path,
    revision,
  }: {
    namespace: string;
    path: string;
    revision: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/?op=delete-revision&ref=${revision}`,
  method: "POST",
  schema: TreeNodeDeletedSchema,
});

export const useDeleteRevision = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path, revision }: { path: string; revision: string }) =>
      deleteRevision({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          path,
          namespace: namespace,
          revision,
        },
      }),
    onSuccess(_, variables) {
      toast({
        title: `revision deleted`,
        description: `revision ${variables.revision.slice(0, 8)} was deleted`,
        variant: "success",
      });
      // deleting a revision, deletes all corresponding tags,
      // since we don't know that relation, we have to invalidate
      // our cache to force a refetch
      queryClient.invalidateQueries(
        treeKeys.tagsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      // since tags are part of the revisions list, it must also be
      // invalidated (and revision we just deleted must be removed as well)
      queryClient.invalidateQueries(
        treeKeys.revisionsList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not delete 😢",
        variant: "error",
      });
    },
  });
};
