import { RoleDeletedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { roleKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteGroup = apiFactory({
  url: ({ namespace, groupName }: { namespace: string; groupName: string }) =>
    `/api/v2/namespaces/${namespace}/groups/${groupName}`,
  method: "DELETE",
  schema: RoleDeletedSchema,
});

export const useDeleteGroup = ({
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
    mutationFn: (groupName: string) =>
      deleteGroup({
        apiKey: apiKey ?? undefined,
        urlParams: { groupName, namespace },
      }),
    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: roleKeys.roleList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.roles.mutate.deleteRole.success.title"),
        description: t("api.roles.mutate.deleteRole.success.description"),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.roles.mutate.deleteRole.error.description"),
        variant: "error",
      });
    },
  });
};
