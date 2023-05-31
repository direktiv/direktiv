import {
  RegistryDeletedSchema,
  RegistryListSchemaType,
  RegistrySchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "../../utils";
import { registriesKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const updateCache = (
  oldData: RegistryListSchemaType | undefined,
  deletedItem: RegistrySchemaType
) => {
  if (!oldData) return undefined;
  const oldRegistries = oldData.registries;

  return {
    ...oldData,
    ...(oldRegistries
      ? {
          registries: oldRegistries.filter(
            (item: RegistrySchemaType) => item.name !== deletedItem.name
          ),
        }
      : {}),
  };
};

const deleteRegistry = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/functions/registries/namespaces/${namespace}`,
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

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ registry }: { registry: RegistrySchemaType }) =>
      deleteRegistry({
        apiKey: apiKey ?? undefined,
        payload: {
          reg: registry.name,
        },
        urlParams: {
          namespace: namespace,
        },
        headers: undefined,
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
        title: "secret deleted",
        description: `tag ${variables.registry.name} was deleted`,
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not delete 😢",
        variant: "error",
      });
    },
  });
};
