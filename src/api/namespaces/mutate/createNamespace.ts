import { useMutation, useQueryClient } from "@tanstack/react-query";

import { NamespaceCreatedSchema } from "../schema";
import type { NamespaceListSchemaType } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { sortByName } from "~/api/tree/utils";
import { useApiKey } from "~/util/store/apiKey";
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

  return useMutation({
    mutationFn: ({ name }: { name: string }) =>
      createNamespace({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        headers: undefined,
        urlParams: {
          name,
        },
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
