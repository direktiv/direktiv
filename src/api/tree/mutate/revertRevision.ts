import { NodeListSchemaType, WorkflowCreatedSchema } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
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

export const useRevertRevision = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error(t("api.tree.mutate.revertRevision.error.undefined"));
  }

  return useMutation({
    mutationFn: ({ path }: { path: string }) =>
      revertRevision({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        urlParams: {
          namespace: namespace,
          path,
        },
      }),
    onSuccess(data, variables) {
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
        title: t("api.tree.mutate.revertRevision.error.title"),
        description: t("api.tree.mutate.revertRevision.error.description"),
        variant: "error",
      });
    },
  });
};
