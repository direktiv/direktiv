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

const createNamespace = apiFactory({
  url: () => "/api/v2/namespaces",
  method: "POST",
  schema: NamespaceCreatedEditedSchema,
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
        urlParams: {},
        payload: {
          name,
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
        title: t("api.namespaces.mutate.create.success.title"),
        description: t("api.namespaces.mutate.create.success.description", {
          name: variables.name,
        }),
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.namespaces.mutate.create.error.description"),
        variant: "error",
      });
    },
  });
};
