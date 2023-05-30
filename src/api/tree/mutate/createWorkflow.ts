import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "~/api/utils";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createWorkflow = apiFactory({
  url: ({
    namespace,
    path,
    name,
  }: {
    namespace: string;
    path?: string;
    name: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/${name}?op=create-workflow`,
  method: "PUT",
  schema: WorkflowCreatedSchema,
});

type ResolvedCreateWorkflow = Awaited<ReturnType<typeof createWorkflow>>;

export const useCreateWorkflow = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateWorkflow) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({
      path,
      name,
      fileContent,
    }: {
      path?: string;
      name: string;
      fileContent: string;
    }) =>
      createWorkflow({
        apiKey: apiKey ?? undefined,
        payload: fileContent,
        urlParams: {
          namespace: namespace,
          path,
          name,
        },
      }),
    onSuccess(data, variables) {
      toast({
        title: t("api.tree.mutate.createWorkflow.success.title"),
        description: t("api.tree.mutate.createWorkflow.success.description", {
          name: variables.name,
          path: variables.path,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.tree.mutate.createWorkflow.error.title"),
        description: t("api.tree.mutate.createWorkflow.error.description"),
        variant: "error",
      });
    },
  });
};
