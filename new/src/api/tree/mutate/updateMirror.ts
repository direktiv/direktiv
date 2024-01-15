import {
  MirrorPostSchemaType,
  UpdateMirrorResponseSchema,
  UpdateMirrorResponseSchemaType,
} from "../schema/mirror";

import { apiFactory } from "~/api/apiFactory";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateMirror = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/tree?op=update-mirror`,
  method: "POST",
  schema: UpdateMirrorResponseSchema,
});

export const useUpdateMirror = ({
  onSuccess,
}: {
  onSuccess?: (data: UpdateMirrorResponseSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  const mutationFn = ({
    name,
    mirror,
  }: {
    name: string;
    mirror: MirrorPostSchemaType;
  }) =>
    updateMirror({
      apiKey: apiKey ?? undefined,
      payload: mirror,
      urlParams: {
        namespace: name,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries(
        treeKeys.mirrorInfo(variables.name, {
          apiKey: apiKey ?? undefined,
        })
      );
      onSuccess?.(data);
      toast({
        title: t("api.namespaces.mutate.updateMirror.success.title"),
        description: t(
          "api.namespaces.mutate.updateMirror.success.description",
          {
            namespace: variables.name,
          }
        ),
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.namespaces.mutate.updateMirror.error.description"),
        variant: "error",
      });
    },
  });
};
