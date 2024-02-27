import { BaseFileSchemaType, FileDeletedSchema } from "../schema";
import {
  forceLeadingSlash,
  getFilenameFromPath,
  getParentFromPath,
} from "~/api/files/utils";

import { apiFactory } from "~/api/apiFactory";
import { fileKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteFile = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/v2/namespaces/${namespace}/files${forceLeadingSlash(path)}`,
  method: "DELETE",
  schema: FileDeletedSchema,
});

export const useDeleteFile = ({
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
    mutationFn: ({ file }: { file: BaseFileSchemaType }) =>
      deleteFile({
        apiKey: apiKey ?? undefined,
        urlParams: {
          path: file.path,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.invalidateQueries(
        fileKeys.file(namespace, {
          apiKey: apiKey ?? undefined,
          path: getParentFromPath(variables.file.path),
        })
      );
      toast({
        title: t("api.tree.mutate.file.delete.success.title"),
        description: t("api.tree.mutate.file.delete.success.description", {
          name: getFilenameFromPath(variables.file.path),
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: (_, variables) => {
      toast({
        title: t("api.generic.error"),
        description: t("api.tree.mutate.file.delete.error.description", {
          name: getFilenameFromPath(variables.file.path),
        }),
        variant: "error",
      });
    },
  });
};
