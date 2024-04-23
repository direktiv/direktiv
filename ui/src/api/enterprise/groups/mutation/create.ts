import { GroupCreatedEditedSchema, GroupFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { groupKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createGroup = apiFactory<GroupFormSchemaType>({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/groups`,
  method: "POST",
  schema: GroupCreatedEditedSchema,
});

type ResolvedCreateGroup = Awaited<ReturnType<typeof createGroup>>;

export const useCreateGroup = ({
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
    mutationFn: (tokenFormProps: GroupFormSchemaType) =>
      createGroup({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data, { description }) {
      queryClient.invalidateQueries({
        queryKey: groupKeys.groupList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.groups.mutate.createGroup.success.title"),
        description: t("api.groups.mutate.createGroup.success.description", {
          name: description,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.groups.mutate.createGroup.error.description"),
        variant: "error",
      });
    },
  });
};
