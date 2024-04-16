import { NamespaceDeletedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteNamespace = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}?recursive=true`,
  method: "DELETE",
  schema: NamespaceDeletedSchema,
});

type ResolvedDeleteNamespace = Awaited<ReturnType<typeof deleteNamespace>>;

export const useDeleteNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedDeleteNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutationWithPermissions({
    mutationFn: ({ namespace }: { namespace: string }) =>
      deleteNamespace({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
      }),
    onSuccess(data, variables) {
      /**
       * invalidating the cache is important here, because after deleting the namespace
       * we will redirect to the frontpage, where we pick the first namespace we can
       * find and redirect to it. It is very likely that the cache will be used here
       * (namespace cache gets populated very ealy in the app lifecycle), so we need
       * to make sure that we don't accidentally redirect to the namespace we just
       * deleted.
       */
      queryClient.invalidateQueries({
        queryKey: namespaceKeys.all(apiKey ?? undefined),
      });
      toast({
        title: t("api.namespaces.mutate.deleteNamespaces.success.title"),
        description: t(
          "api.namespaces.mutate.deleteNamespaces.success.description",
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
          "api.namespaces.mutate.deleteNamespaces.error.description"
        ),
        variant: "error",
      });
    },
  });
};
