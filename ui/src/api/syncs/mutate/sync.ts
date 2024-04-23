import {
  SyncListSchemaType,
  SyncResponseSchema,
  SyncResponseSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { syncKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createSync = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}/syncs`,
  method: "POST",
  schema: SyncResponseSchema,
});

const updateCache = (
  oldData: SyncListSchemaType | undefined,
  newData: SyncResponseSchemaType
) => {
  const newRecord = newData.data;
  if (!oldData) return { data: [newRecord] };

  const oldRecords = oldData.data;
  return {
    data: [...oldRecords, newRecord],
  };
};

export const useSync = ({
  onSuccess,
}: {
  onSuccess?: (data: SyncResponseSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const mutationFn = ({ namespace }: { namespace: string | null }) => {
    if (!namespace) {
      throw new Error("namespace is undefined");
    }
    return createSync({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
      },
    });
  };

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      if (!variables.namespace) {
        throw new Error("variables.namespace is undefined");
      }
      queryClient.setQueryData<SyncListSchemaType>(
        syncKeys.syncsList(variables.namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, data)
      );
      onSuccess?.(data);
      toast({
        title: t("api.namespaces.mutate.syncMirror.success.title"),
        description: t("api.namespaces.mutate.syncMirror.success.description", {
          namespace: variables.namespace,
        }),
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.namespaces.mutate.syncMirror.error.description"),
        variant: "error",
      });
    },
  });
};
