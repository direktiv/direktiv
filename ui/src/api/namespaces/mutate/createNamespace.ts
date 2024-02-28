import { MirrorPostSchemaType } from "~/api/tree/schema/mirror";
import { NamespaceCreatedSchema } from "../schema";
import type { NamespaceListSchemaType } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { sortByName } from "~/api/files/utils";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const createNamespace = apiFactory({
  url: ({ name }: { name: string }) => `/api/namespaces/${name}`,
  method: "PUT",
  schema: NamespaceCreatedSchema,
});

type ResolvedCreateNamespace = Awaited<ReturnType<typeof createNamespace>>;

export const useCreateNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutationWithPermissions({
    mutationFn: ({
      name,
      mirror,
    }: {
      name: string;
      mirror?: MirrorPostSchemaType;
    }) =>
      createNamespace({
        apiKey: apiKey ?? undefined,
        urlParams: {
          name,
        },
        payload: mirror,
      }),
    onSuccess(data, variables) {
      queryClient.setQueryData<NamespaceListSchemaType>(
        namespaceKeys.all(apiKey ?? undefined),
        (oldData) => {
          if (!oldData) return undefined;
          const oldResults = oldData?.results;
          return {
            ...oldData,
            results: [...oldResults, data.namespace].sort(sortByName),
          };
        }
      );
      toast({
        title: t("api.namespaces.mutate.createNamespaces.success.title"),
        description: t(
          "api.namespaces.mutate.createNamespaces.success.description",
          { name: variables.name }
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
