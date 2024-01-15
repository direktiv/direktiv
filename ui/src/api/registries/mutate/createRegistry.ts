import { RegistryCreatedSchema, RegistryFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { registriesKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

export const createRegistry = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/registries`,
  method: "POST",
  schema: RegistryCreatedSchema,
});

export const useCreateRegistry = ({
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

  const mutationFn = ({ url, user, password }: RegistryFormSchemaType) =>
    createRegistry({
      apiKey: apiKey ?? undefined,
      payload: { user, password, url },
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (registry, variables) => {
      queryClient.invalidateQueries(
        registriesKeys.registriesList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.registries.mutate.createRegistry.success.title"),
        description: t(
          "api.registries.mutate.createRegistry.success.description",
          { name: variables.url }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.registries.mutate.createRegistry.error.description"
        ),
        variant: "error",
      });
    },
  });
};
