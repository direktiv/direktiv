import {
  RevisionDeletedSchema,
  RevisionsListSchemaType,
} from "../schema/revisions";

import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: RevisionsListSchemaType | undefined,
  variables: Parameters<
    ReturnType<typeof useDeleteServiceRevision>["mutate"]
  >[0]
) => {
  if (!oldData) return undefined;
  const remainingRevisions = (oldData.revisions ?? []).filter(
    (service) => service.revision !== variables.revision
  );
  return {
    ...oldData,
    revisions: remainingRevisions,
  };
};

const deleteServiceRevision = apiFactory({
  url: ({
    namespace,
    service,
    revision,
  }: {
    namespace: string;
    service: string;
    revision: string;
  }) =>
    `/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}`,
  method: "DELETE",
  schema: RevisionDeletedSchema,
});

export const useDeleteServiceRevision = ({
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
    mutationFn: ({
      service,
      revision,
    }: {
      service: string;
      revision: string;
    }) =>
      deleteServiceRevision({
        apiKey: apiKey ?? undefined,
        urlParams: {
          service,
          namespace,
          revision,
        },
      }),
    onSuccess(_, variables) {
      queryClient.setQueryData<RevisionsListSchemaType>(
        serviceKeys.serviceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          service: variables.service,
        }),
        (oldData) => updateCache(oldData, variables)
      );
      toast({
        title: t("api.services.mutate.deleteServiceRevision.success.title"),
        description: t(
          "api.services.mutate.deleteServiceRevision.success.description",
          {
            name: variables.revision,
          }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.services.mutate.deleteServiceRevision.error.description"
        ),
        variant: "error",
      });
    },
  });
};
