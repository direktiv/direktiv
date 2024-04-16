import { MirrorPostSchemaType } from "~/api/tree/schema/mirror";
import { NamespaceCreatedEditedSchema } from "../schema";
import type { NamespaceListSchemaType } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { sortByName } from "~/api/files/utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const updateNamespace = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}`,
  method: "PATCH",
  schema: NamespaceCreatedEditedSchema,
});

type ResolvedUpdateNamespace = Awaited<ReturnType<typeof updateNamespace>>;

export const useUpdateNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedUpdateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutationWithPermissions({
    mutationFn: ({
      namespace,
      mirror,
    }: {
      namespace: string;
      mirror?: MirrorPostSchemaType;
    }) =>
      updateNamespace({
        apiKey: apiKey ?? undefined,
        urlParams: { namespace },
        payload: {
          mirror,
        },
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<NamespaceListSchemaType>(
        namespaceKeys.all(apiKey ?? undefined),
        (oldData) => {
          if (!oldData) return undefined;
          const oldResults = oldData?.data;
          return {
            ...oldData,
            results: [...oldResults, data.data].sort(sortByName),
          };
        }
      );
      toast({
        title: t("api.namespaces.mutate.createNamespaces.success.title"),
        description: t(
          "api.namespaces.mutate.createNamespaces.success.description",
          { name: variables.namespace }
        ),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t(
          "api.namespaces.mutate.createNamespaces.error.description"
        ),
        variant: "error",
      });
    },
  });
};
