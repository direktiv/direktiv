import { GroupCreatedEditedSchema, GroupFormSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { roleKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createRole = apiFactory<GroupFormSchemaType>({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/groups`,
  method: "POST",
  schema: GroupCreatedEditedSchema,
});

type ResolvedCreateRole = Awaited<ReturnType<typeof createRole>>;

export const useCreateRole = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateRole) => void } = {}) => {
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
      createRole({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
        payload: tokenFormProps,
      }),
    onSuccess(data, { description }) {
      queryClient.invalidateQueries({
        queryKey: roleKeys.roleList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
      });
      toast({
        title: t("api.roles.mutate.createRole.success.title"),
        description: t("api.roles.mutate.createRole.success.description", {
          name: description,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.roles.mutate.createRole.error.description"),
        variant: "error",
      });
    },
  });
};
