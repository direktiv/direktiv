import { SecretUpdatedSchema, SecretUpdatedSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { secretKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const updateSecret = apiFactory({
  url: ({
    baseUrl,
    namespace,
    name,
  }: {
    namespace: string;
    name: string;
    baseUrl?: string;
  }) => `${baseUrl ?? ""}/api/namespaces/${namespace}/secrets/${name}`,
  method: "PUT",
  schema: SecretUpdatedSchema,
});

export const useUpdateSecret = ({
  onSuccess,
}: {
  onSuccess?: (secret: SecretUpdatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ name, value }: { name: string; value: string }) =>
      updateSecret({
        apiKey: apiKey ?? undefined,
        payload: value,
        urlParams: {
          namespace,
          name,
        },
      }),
    onSuccess: (secret) => {
      queryClient.invalidateQueries(
        secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.secrets.mutate.updateSecret.success.title"),
        description: t("api.secrets.mutate.updateSecret.success.description", {
          name: secret.key,
        }),
        variant: "success",
      });
      onSuccess?.(secret);
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
