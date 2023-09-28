import {
  ServiceCreatedSchema,
  ServiceFormSchemaType,
} from "../schema/services";

import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createService = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/functions/namespaces/${namespace}`,
  method: "POST",
  schema: ServiceCreatedSchema,
});

type ResolvedCreateNamespace = Awaited<ReturnType<typeof createService>>;

export const useCreateService = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: (serviceFormProps: ServiceFormSchemaType) =>
      createService({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
        payload: serviceFormProps,
      }),
    onSuccess(data, { name }) {
      queryClient.invalidateQueries(
        serviceKeys.servicesList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.services.mutate.createService.success.title"),
        description: t(
          "api.services.mutate.createService.success.description",
          { name }
        ),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.services.mutate.createService.error.description"),
        variant: "error",
      });
    },
  });
};
