import { ServiceRebuildSchema } from "../schema/services";
import { apiFactory } from "../../apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "../../../util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "../../../design/Toast";
import { useTranslation } from "react-i18next";

const rebuildService = apiFactory({
  url: ({
    baseUrl,
    namespace,
    service,
  }: {
    baseUrl?: string;
    namespace: string;
    service: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/services/${service}/actions/rebuild`,
  method: "POST",
  schema: ServiceRebuildSchema,
});

export const useRebuildService = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: (service: string) =>
      rebuildService({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          service,
        },
      }),
    onSuccess() {
      queryClient.invalidateQueries(
        serviceKeys.servicesList(namespace, { apiKey: apiKey ?? undefined })
      );
      toast({
        title: t("api.services.mutate.rebuildService.success.title"),
        description: t(
          "api.services.mutate.rebuildService.success.description"
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.services.mutate.rebuildService.error.description"),
        variant: "error",
      });
    },
  });
};
