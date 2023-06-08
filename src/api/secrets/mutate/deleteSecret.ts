import {
  SecretDeletedSchema,
  SecretListSchemaType,
  SecretSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "../../utils";
import { secretKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: SecretListSchemaType | undefined,
  deletedItem: SecretSchemaType
) => {
  if (!oldData) return undefined;
  const oldSecrets = oldData.secrets.results;

  return {
    ...oldData,
    ...(oldSecrets
      ? {
          secrets: {
            results: oldSecrets.filter(
              (item: SecretSchemaType) => item.name !== deletedItem.name
            ),
          },
        }
      : {}),
  };
};

const deleteSecret = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/secrets/${name}`,
  method: "DELETE",
  schema: SecretDeletedSchema,
});

export const useDeleteSecret = ({
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

  return useMutation({
    mutationFn: ({ secret }: { secret: SecretSchemaType }) =>
      deleteSecret({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        urlParams: {
          name: secret.name,
          namespace,
        },
        headers: undefined,
      }),
    onSuccess(_, variables) {
      const deletedItem = variables.secret;
      queryClient.setQueryData<SecretListSchemaType>(
        secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, deletedItem)
      );
      toast({
        title: t("api.secrets.mutate.deleteSecret.success.title"),
        description: t("api.secrets.mutate.deleteSecret.success.description", {
          name: variables.secret.name,
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.secrets.mutate.deleteSecret.error.description"),
        variant: "error",
      });
    },
  });
};
