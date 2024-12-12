import { BaseFileSchemaType, FileDeletedSchema } from "../schema";
import { getFilenameFromPath, getParentFromPath } from "~/api/files/utils";

import { apiFactory } from "~/api/apiFactory";
import { fileKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const deleteMultipleFiles = apiFactory({
  url: ({
    baseUrl,
    namespace,
    paths,
  }: {
    baseUrl?: string;
    namespace: string;
    paths: string[];
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/files?ids=${paths.join(
      ","
    )}`,
  method: "DELETE",
  schema: FileDeletedSchema,
});

export const useDeleteMultipleFiles = ({
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
    mutationFn: ({ files }: { files: BaseFileSchemaType[] }) =>
      deleteMultipleFiles({
        apiKey: apiKey ?? undefined,
        urlParams: {
          paths: files.map((file) => file.path),
          namespace,
        },
      }),
    async onSuccess(_, variables) {
      for (const file of variables.files) {
        const selfKey = fileKeys.file(namespace, {
          apiKey: apiKey ?? undefined,
          path: file.path,
        });
        const parentKey = fileKeys.file(namespace, {
          apiKey: apiKey ?? undefined,
          path: getParentFromPath(file.path),
        });
        await queryClient.removeQueries({
          queryKey: selfKey,
        });
        await queryClient.invalidateQueries({
          queryKey: parentKey,
        });

        toast({
          title: t("api.tree.mutate.file.delete.success.title"),
          description: t("api.tree.mutate.file.delete.success.description", {
            name: getFilenameFromPath(file.path),
          }),
          variant: "success",
        });
      }
      onSuccess?.();
    },
    onError: (_, variables) => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.file.delete.error.description", {
          name: getFilenameFromPath(variables.files[0]?.path ?? ""),
        }),
        variant: "error",
      });
    },
  });
};
