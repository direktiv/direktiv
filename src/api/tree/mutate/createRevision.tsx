import { ToastAction, useToast } from "~/design/Toast";

import { WorkflowCreatedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

export const createRevision = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path: string;
  }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=save-workflow&ref=latest`,
  method: "POST",
  schema: WorkflowCreatedSchema,
});

export const useCreateRevision = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({
      path,
    }: {
      path: string;
      createLink: (revision: string) => string;
    }) =>
      createRevision({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess: (data, variables) => {
      toast({
        title: t("api.tree.mutate.createRevision.success.title"),
        description: t("api.tree.mutate.createRevision.success.description", {
          name: data.revision.name,
        }),
        variant: "success",
        action: (
          <ToastAction
            data-testid="make-revision-toast-success-action"
            altText={t("api.tree.mutate.createRevision.success.action")}
            onClick={() => {
              navigate(variables.createLink(data.revision.name));
            }}
          >
            {t("api.tree.mutate.createRevision.success.action")}
          </ToastAction>
        ),
      });
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.createRevision.error.description"),
        variant: "error",
      });
    },
  });
};
