import { RoleCreatedEditedSchema, RoleFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { roleKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const editRole = apiFactory<RoleFormSchemaType>({
  url: ({
    namespace,
    baseUrl,
    roleName,
  }: {
    baseUrl?: string;
    namespace: string;
    roleName: string;
  }) => `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/roles/${roleName}`,
  method: "PUT",
  schema: RoleCreatedEditedSchema,
});

export const useEditRole = ({ onSuccess }: { onSuccess?: () => void } = {}) => {
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
      roleName,
      payload,
    }: {
      roleName: string;
      payload: RoleFormSchemaType;
    }) =>
      editRole({
        apiKey: apiKey ?? undefined,
        urlParams: { roleName, namespace },
        payload,
      }),
    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: roleKeys.roleList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.roles.mutate.editRole.success.title"),
        description: t("api.roles.mutate.editRole.success.description"),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.roles.mutate.editRole.error.description"),
        variant: "error",
      });
    },
  });
};
