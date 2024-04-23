import {
  SecretCreatedUpdatedSchema,
  SecretFormCreateEditSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { secretKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

type CreateSecretParams = { baseUrl?: string; namespace: string };

export const createSecret = apiFactory<SecretFormCreateEditSchemaType>({
  url: ({ baseUrl, namespace }: CreateSecretParams) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/secrets`,
  method: "POST",
  schema: SecretCreatedUpdatedSchema,
});

export const useCreateSecret = ({
  onSuccess,
}: {
  onSuccess?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (payload: SecretFormCreateEditSchemaType) =>
    createSecret({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (secret) => {
      queryClient.invalidateQueries(
        secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.secrets.mutate.updateSecret.success.title"),
        description: t("api.secrets.mutate.updateSecret.success.description", {
          name: secret.data.name,
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.secrets.mutate.updateSecret.error.description"),
        variant: "error",
      });
    },
  });
};
