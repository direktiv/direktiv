import {
  MirrorSyncResponseSchema,
  MirrorSyncResponseSchemaType,
} from "../schema/mirror";

import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const syncMirror = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/tree?op=sync-mirror&force=true`,
  method: "POST",
  schema: MirrorSyncResponseSchema,
});

export const useSyncMirror = ({
  onSuccess,
}: {
  onSuccess?: (data: MirrorSyncResponseSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = () =>
    syncMirror({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data) => {
      queryClient.invalidateQueries(
        treeKeys.mirrorInfo(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      onSuccess?.(data);
      toast({
        title: t("api.namespaces.mutate.syncMirror.success.title"),
        description: t("api.namespaces.mutate.syncMirror.success.description", {
          namespace,
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
