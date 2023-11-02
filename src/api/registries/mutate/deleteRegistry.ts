import {
  RegistryDeletedSchema,
  RegistryListSchemaType,
  RegistrySchemaType,
} from "../schema";

import { apiFactory } from "../../apiFactory";
import { registriesKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "../../../util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "../../../design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: RegistryListSchemaType | undefined,
  deletedItem: RegistrySchemaType
) => {
  if (!oldData) return undefined;
  const oldRegistries = oldData.data;
  return {
    ...oldData,
    ...(oldRegistries
      ? {
          data: oldRegistries.filter(
            (item: RegistrySchemaType) => item.id !== deletedItem.id
          ),
        }
      : {}),
  };
};

const deleteRegistry = apiFactory({
  url: ({ namespace, gegistryId }: { namespace: string; gegistryId: string }) =>
    `/api/v2/namespaces/${namespace}/registries/${gegistryId}`,
  method: "DELETE",
  schema: RegistryDeletedSchema,
});

export const useDeleteRegistry = ({
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
    mutationFn: ({ registry }: { registry: RegistrySchemaType }) =>
      deleteRegistry({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          gegistryId: registry.id,
        },
      }),
    onSuccess(_, variables) {
      const deletedItem = variables.registry;
      queryClient.setQueryData<RegistryListSchemaType>(
        registriesKeys.registriesList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, deletedItem)
      );
      toast({
        title: t("api.registries.mutate.deleteRegistry.success.title"),
        description: t(
          "api.registries.mutate.deleteRegistry.success.description",
          { name: variables.registry.url }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.registries.mutate.deleteRegistry.error.description"
        ),
        variant: "error",
      });
    },
  });
};
