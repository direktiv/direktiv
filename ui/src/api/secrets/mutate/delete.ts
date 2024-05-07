import { SecretSchemaType, SecretsDeletedSchema } from "../schema";

import { apiFactory } from "../../apiFactory";
import { secretKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "../../../util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "../../../design/Toast";
import { useTranslation } from "react-i18next";

type DeleteSecretParams = { namespace: string; name: string };

const deleteSecret = apiFactory({
  url: ({ namespace, name }: DeleteSecretParams) =>
    `/api/v2/namespaces/${namespace}/secrets/${name}`,
  method: "DELETE",
  schema: SecretsDeletedSchema,
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

  return useMutationWithPermissions({
    mutationFn: ({ secret }: { secret: SecretSchemaType }) =>
      deleteSecret({
        apiKey: apiKey ?? undefined,
        urlParams: {
          name: secret.name,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.invalidateQueries({
        queryKey: secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
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
