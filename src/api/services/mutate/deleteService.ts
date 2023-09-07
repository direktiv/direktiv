import {
  ServiceDeletedSchema,
  ServicesListSchemaType,
} from "../schema/services";

import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useMutationWithPermissionHandling } from "~/api/errorHandling";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: ServicesListSchemaType | undefined,
  variables: Parameters<ReturnType<typeof useDeleteService>["mutate"]>[0]
) => {
  if (!oldData) return undefined;
  const remainingFunctions = oldData.functions.filter(
    (service) => service.info.name !== variables.service
  );
  return {
    ...oldData,
    functions: remainingFunctions,
  };
};

const deleteService = apiFactory({
  url: ({ namespace, service }: { namespace: string; service: string }) =>
    `/api/functions/namespaces/${namespace}/function/${service}`,
  method: "DELETE",
  schema: ServiceDeletedSchema,
});

export const useDeleteService = ({
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

  return useMutationWithPermissionHandling({
    mutationFn: ({ service }: { service: string }) =>
      deleteService({
        apiKey: apiKey ?? undefined,
        urlParams: {
          service,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.setQueryData<ServicesListSchemaType>(
        serviceKeys.servicesList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, variables)
      );
      toast({
        title: t("api.services.mutate.deleteService.success.title"),
        description: t(
          "api.services.mutate.deleteService.success.description",
          {
            name: variables.service,
          }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.services.mutate.deleteService.error.description"),
        variant: "error",
      });
    },
  });
};
