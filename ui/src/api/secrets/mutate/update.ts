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

type UpdateSecretParams = { baseUrl?: string; namespace: string; name: string };

export const updateSecret = apiFactory({
  url: ({ baseUrl, namespace, name }: UpdateSecretParams) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/secrets/${name}`,
  method: "PATCH",
  schema: SecretCreatedUpdatedSchema,
});

export const useUpdateSecret = ({
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

  const mutationFn = ({ name, ...payload }: SecretFormCreateEditSchemaType) =>
    updateSecret({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        name,
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (secret) => {
      queryClient.invalidateQueries({
        queryKey: secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
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
