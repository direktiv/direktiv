import {
  RevisionCreatedSchema,
  RevisionFormSchemaType,
} from "../schema/revisions";

import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createServiceRevision = apiFactory({
  url: ({ namespace, service }: { namespace: string; service: string }) =>
    `/api/functions/namespaces/${namespace}/function/${service}`,
  method: "POST",
  schema: RevisionCreatedSchema,
});

type ResolvedCreateNamespace = Awaited<
  ReturnType<typeof createServiceRevision>
>;

export const useCreateServiceRevision = ({
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
    mutationFn: ({
      service,
      payload,
    }: {
      service: string;
      payload: RevisionFormSchemaType;
    }) =>
      createServiceRevision({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          service,
        },
        payload,
      }),
    onSuccess(data, variables) {
      queryClient.invalidateQueries(
        serviceKeys.serviceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          service: variables.service,
        })
      );
      toast({
        title: t("api.services.mutate.createServiceRevision.success.title"),
        description: t(
          "api.services.mutate.createServiceRevision.success.description"
        ),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.services.mutate.createServiceRevision.error.description"
        ),
        variant: "error",
      });
    },
  });
};
