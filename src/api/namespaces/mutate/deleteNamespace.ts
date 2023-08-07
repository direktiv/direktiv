import { NamespaceDeletedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteNamespace = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}?recursive=true`,
  method: "DELETE",
  schema: NamespaceDeletedSchema,
});

type ResolvedCreateNamespace = Awaited<ReturnType<typeof deleteNamespace>>;

export const useDeleteNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: ({ namespace }: { namespace: string }) =>
      deleteNamespace({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
        },
      }),
    onSuccess(data, variables) {
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
