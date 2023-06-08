import {
  SecretCreatedSchema,
  SecretCreatedSchemaType,
  SecretListSchemaType,
  SecretSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { secretKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: SecretListSchemaType | undefined,
  createdItem: SecretCreatedSchemaType
) => {
  if (!oldData) return undefined;
  const newListItem: SecretSchemaType = { name: createdItem.key };
  const oldResults = oldData.secrets.results;
  return {
    ...oldData,
    secrets: {
      results: [...oldResults, newListItem].sort((a, b) =>
        a.name.localeCompare(b.name)
      ),
    },
  };
};

const createSecret = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/secrets/${name}`,
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
        headers: undefined,
      }),
    onSuccess: (secret) => {
      queryClient.setQueryData<SecretListSchemaType>(
        secretKeys.secretsList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, secret)
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
        title: "api.generic.error",
        description: t("api.secrets.mutate.createSecret.error.description"),
        variant: "error",
      });
    },
  });
};
