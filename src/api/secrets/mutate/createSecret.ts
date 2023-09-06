import { SecretCreatedSchema, SecretCreatedSchemaType } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { secretKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const createSecret = apiFactory({
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
  schema: SecretCreatedSchema,
});

export const useCreateSecret = ({
  onSuccess,
}: {
  onSuccess?: (secret: SecretCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ name, value }: { name: string; value: string }) =>
      createSecret({
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
        title: t("api.secrets.mutate.createSecret.success.title"),
        description: t("api.secrets.mutate.createSecret.success.description", {
          name: secret.key,
        }),
        variant: "success",
      });
      onSuccess?.(secret);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.secrets.mutate.createSecret.error.description"),
        variant: "error",
      });
    },
  });
};
