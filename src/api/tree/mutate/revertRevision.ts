import { NodeListSchemaType, WorkflowCreatedSchema } from "../schema/node";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const revertRevision = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=discard-workflow&ref=latest`,
  method: "POST",
  schema: WorkflowCreatedSchema,
});

export const useRevertRevision = ({
  onSuccess,
}: {
  onSuccess?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ path }: { path: string }) =>
      revertRevision({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess(data, variables) {
      onSuccess?.();
      queryClient.setQueryData<NodeListSchemaType>(
        treeKeys.nodeContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        () => data
      );
      toast({
        title: t("api.tree.mutate.revertRevision.success.title"),
        description: t("api.tree.mutate.revertRevision.success.description"),
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.revertRevision.error.description"),
        variant: "error",
      });
    },
  });
};
