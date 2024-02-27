import { CreateFileSchemaType, FileCreatedSchema } from "../schema";
import {
  forceLeadingSlash,
  getFilenameFromPath,
  getParentFromPath,
} from "../utils";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const createFile = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path?: string;
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/files${forceLeadingSlash(
      path
    )}`,
  method: "POST",
  schema: FileCreatedSchema,
});

type ResolvedCreateFile = Awaited<ReturnType<typeof createFile>>;

export const useCreateFile = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateFile) => void } = {}) => {
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
      payload,
    }: {
      path?: string;
      payload: CreateFileSchemaType;
    }) =>
      createFile({
        apiKey: apiKey ?? undefined,
        payload,
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess(data, variables) {
      const fileType =
        variables.payload.type === "directory" ? "directory" : "file";
      toast({
        title: t(`api.tree.mutate.${fileType}.create.success.title`),
        description: t(`api.tree.mutate.file.create.success.description`, {
          name: getFilenameFromPath(variables.payload.name),
          path: getParentFromPath(data.data.path),
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: (_, variables) => {
      toast({
        title: t("api.generic.error"),
        description: t(`api.tree.mutate.file.create.error.description`, {
          name: getFilenameFromPath(variables.payload.name),
        }),
        variant: "error",
      });
    },
  });
};
