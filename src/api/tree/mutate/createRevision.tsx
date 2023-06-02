import { ToastAction, useToast } from "~/design/Toast";

import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "~/api/utils";
import { forceLeadingSlash } from "../utils";
import { pages } from "~/util/router/pages";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
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

  return useMutation({
    mutationFn: ({ path }: { path: string }) =>
      createRevision({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        headers: undefined,
        urlParams: {
          namespace: namespace,
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
            altText="Open Revision"
            onClick={() => {
              navigate(
                pages.explorer.createHref({
                  namespace,
                  path: variables.path,
                  subpage: "workflow-revisions",
                  revision: data.revision.name,
                })
              );
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
