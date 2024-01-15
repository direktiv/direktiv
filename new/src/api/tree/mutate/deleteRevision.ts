import { NodeDeletedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteRevision = apiFactory({
  url: ({
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
  schema: NodeDeletedSchema,
});

export const useDeleteRevision = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ path, revision }: { path: string; revision: string }) =>
      deleteRevision({
        apiKey: apiKey ?? undefined,
        urlParams: {
          path,
          namespace,
          revision,
        },
      }),
    onSuccess(_, variables) {
      toast({
        title: t("api.tree.mutate.deleteRevision.success.title"),
        description: t("api.tree.mutate.deleteRevision.success.description", {
          name: variables.revision,
        }),
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
        title: t("api.generic.error"),
        description: t("api.tree.mutate.deleteRevision.error.description"),
        variant: "error",
      });
    },
  });
};
