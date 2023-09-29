import {
  ServiceDeletedSchema,
  ServicesListSchemaType,
} from "../schema/services";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
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
  url: ({
    namespace,
    service,
    workflow,
    version,
  }: {
    namespace: string;
    service: string;
    workflow?: string;
    version?: string;
  }) =>
    workflow && version
      ? `/api/functions/namespaces/${namespace}/tree${forceLeadingSlash(
          workflow
        )}?op=delete-service&svn=${service}&version=${version}`
      : `/api/functions/namespaces/${namespace}/function/${service}`,
  method: "DELETE",
  schema: ServiceDeletedSchema,
});

export const useDeleteService = ({
  workflow,
  version,
  onSuccess,
}: { workflow?: string; version?: string; onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({
      service,
    }: {
      service: string;
      workflow?: string;
      version?: string;
    }) =>
      deleteService({
        apiKey: apiKey ?? undefined,
        urlParams: {
          service,
          namespace,
          workflow,
          version,
        },
      }),
    onSuccess(_, variables) {
      queryClient.setQueryData<ServicesListSchemaType>(
        serviceKeys.servicesList(namespace, {
          apiKey: apiKey ?? undefined,
          workflow,
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
