import {
  GroupDeletedSchema,
  GroupSchemaType,
  GroupsListSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { groupKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";

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

// const deleteGroup = apiFactory({
//   url: ({ namespace, groupId }: { namespace: string; groupId: string }) =>
//     `/api/v2/namespaces/${namespace}/groups/${groupId}`,
//   method: "DELETE",
//   schema: GroupDeletedSchema,
// });

// TODO: remove this mock
const deleteGroup = (_params: {
  apiKey?: string;
  urlParams: { namespace: string; groupId: string };
}): Promise<z.infer<typeof GroupDeletedSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve(null);
    }, 500);
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

  return useMutation({
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
