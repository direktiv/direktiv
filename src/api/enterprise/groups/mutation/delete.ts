import {
  GroupDeletedSchema,
  GroupSchemaType,
  GroupsListSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { groupKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateCache = (
  oldData: GroupsListSchemaType | undefined,
  variables: Parameters<ReturnType<typeof useDeleteGroup>["mutate"]>[0]
) => {
  if (!oldData) return undefined;
  const remainingGroups = oldData.groups.filter(
    (group) => group.id !== variables.id
  );
  return {
    ...oldData,
    groups: remainingGroups,
  };
};

const deleteGroup = apiFactory({
  url: ({ namespace, groupId }: { namespace: string; groupId: string }) =>
    `/api/v2/namespaces/${namespace}/groups/${groupId}`,
  method: "DELETE",
  schema: GroupDeletedSchema,
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
    mutationFn: (group: GroupSchemaType) =>
      deleteGroup({
        apiKey: apiKey ?? undefined,
        urlParams: {
          groupId: group.id,
          namespace,
        },
      }),
    onSuccess(_, variables) {
      queryClient.setQueryData<GroupsListSchemaType>(
        groupKeys.groupList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, variables)
      );
      toast({
        title: t("api.groups.mutate.deleteGroup.success.title"),
        description: t("api.groups.mutate.deleteGroup.success.description", {
          name: variables.group,
        }),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.groups.mutate.deleteGroup.error.description"),
        variant: "error",
      });
    },
  });
};
