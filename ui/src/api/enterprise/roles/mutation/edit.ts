import { GroupCreatedEditedSchema, GroupFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { roleKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const editRole = apiFactory<GroupFormSchemaType>({
  url: ({
    namespace,
    baseUrl,
    groupId,
  }: {
    baseUrl?: string;
    namespace: string;
    groupId: string;
  }) => `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/groups/${groupId}`,
  method: "PUT",
  schema: GroupCreatedEditedSchema,
});

type ResolvedCreateGroup = Awaited<ReturnType<typeof editRole>>;

export const useEditRole = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateGroup) => void } = {}) => {
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
      groupId,
      tokenFormProps,
    }: {
      groupId: string;
      tokenFormProps: GroupFormSchemaType;
    }) =>
      editRole({
        apiKey: apiKey ?? undefined,
        urlParams: {
          groupId,
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data, { tokenFormProps: { description } }) {
      queryClient.invalidateQueries({
        queryKey: roleKeys.roleList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.roles.mutate.editRole.success.title"),
        description: t("api.roles.mutate.editRole.success.description", {
          name: description,
        }),
        variant: "success",
      });
      onSuccess?.(data);
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
