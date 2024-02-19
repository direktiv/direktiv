import { WorkflowCreatedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const createWorkflow = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
    name,
  }: {
    baseUrl?: string;
    namespace: string;
    path?: string;
    name: string;
  }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/tree${forceLeadingSlash(
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

  return useMutationWithPermissions({
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
          namespace,
          path,
          name,
        },
      }),
    onSuccess(data, variables) {
      toast({
        title: t("api.tree.mutate.file.create.success.title"),
        description: t("api.tree.mutate.file.create.success.description", {
          name: variables.name,
          path: variables.path,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.file.create.error.description"),
        variant: "error",
      });
    },
  });
};
