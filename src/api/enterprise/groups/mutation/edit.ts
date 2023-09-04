import { GroupCreatedEditedSchema, GroupFormSchemaType } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { groupKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";

// TODO: remove the line below and delete the mock function
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const editGroup = apiFactory({
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

const editGroupMock = (_params: {
  apiKey?: string;
  urlParams: { namespace: string; groupId: string };
  payload: GroupFormSchemaType;
}): Promise<z.infer<typeof GroupCreatedEditedSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        id: _params.urlParams.groupId,
      });
    }, 500);
  });

type ResolvedCreateGroup = Awaited<ReturnType<typeof editGroup>>;

export const useEditGroup = ({
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

  return useMutation({
    mutationFn: ({
      groupId,
      tokenFormProps,
    }: {
      groupId: string;
      tokenFormProps: GroupFormSchemaType;
    }) =>
      editGroupMock({
        apiKey: apiKey ?? undefined,
        urlParams: {
          groupId,
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data, { tokenFormProps: { description } }) {
      queryClient.invalidateQueries(
        groupKeys.groupList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.groups.mutate.editGroup.success.title"),
        description: t("api.groups.mutate.editGroup.success.description", {
          name: description,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.groups.mutate.editGroup.error.description"),
        variant: "error",
      });
    },
  });
};
