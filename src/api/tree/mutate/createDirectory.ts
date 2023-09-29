import { FolderCreatedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createDirectory = apiFactory({
  url: ({
    namespace,
    path,
    directory,
  }: {
    namespace: string;
    path?: string;
    directory: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/${directory}?op=create-directory`,
  method: "PUT",
  schema: FolderCreatedSchema,
});

type ResolvedCreateDirectory = Awaited<ReturnType<typeof createDirectory>>;

export const useCreateDirectory = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateDirectory) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ path, directory }: { path?: string; directory: string }) =>
      createDirectory({
        apiKey: apiKey ?? undefined,
        urlParams: {
          directory,
          namespace,
          path,
        },
      }),
    onSuccess(data, variables) {
      toast({
        title: t("api.tree.mutate.createDirectory.success.title"),
        description: t("api.tree.mutate.createDirectory.success.description", {
          directory: variables.directory,
          path: variables.path,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.createDirectory.error.description"),
        variant: "error",
      });
    },
  });
};
