import type {
  NamespaceCreatedEditedSchemaType,
  NamespaceListSchemaType,
} from "../schema/namespace";

import { MirrorPostPatchSchemaType } from "~/api/namespaces/schema/mirror";
import { NamespaceCreatedEditedSchema } from "../schema/namespace";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateNamespace = apiFactory<{ mirror?: MirrorPostPatchSchemaType }>({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}`,
  method: "PATCH",
  schema: NamespaceCreatedEditedSchema,
});

type ResolvedUpdateNamespace = Awaited<ReturnType<typeof updateNamespace>>;

const updateCache = (
  oldData: NamespaceListSchemaType | undefined,
  newData: NamespaceCreatedEditedSchemaType
) => {
  if (!oldData) return undefined;
  const newRecord = newData.data;
  const oldRecords = oldData?.data;
  const newRecords = oldRecords.map((record) =>
    record.name === newRecord.name ? newRecord : record
  );
  return {
    data: newRecords,
  };
};

export const useUpdateNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedUpdateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutationWithPermissions({
    mutationFn: ({
      namespace,
      mirror,
    }: {
      namespace: string;
      mirror?: MirrorPostPatchSchemaType;
    }) =>
      updateNamespace({
        apiKey: apiKey ?? undefined,
        urlParams: { namespace },
        payload: {
          mirror,
        },
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<NamespaceListSchemaType>(
        namespaceKeys.all(apiKey ?? undefined),
        (oldData) => updateCache(oldData, data)
      );
      toast({
        title: t("api.namespaces.mutate.update.success.title"),
        description: t("api.namespaces.mutate.update.success.description", {
          name: variables.namespace,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.namespaces.mutate.update.error.description"),
        variant: "error",
      });
    },
  });
};
