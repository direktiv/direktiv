import { useMutation, useQueryClient } from "@tanstack/react-query";

import { NamespaceCreatedSchema } from "../schema";
import type { NamespaceListSchemaType } from "../schema";
import { apiFactory } from "../../utils";
import { namespaceKeys } from "..";
import { sortByName } from "../../tree/utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useToast } from "../../../design/Toast";

const createNamespace = apiFactory({
  pathFn: ({ name }: { name: string }) => `/api/namespaces/${name}`,
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

  return useMutation({
    mutationFn: ({ name }: { name: string }) =>
      createNamespace({
        apiKey: apiKey ?? undefined,
        payload: undefined,
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
        title: "Namespace created",
        description: `Namespace ${variables.name} was created successfully.`,
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create namespace 😢",
        variant: "error",
      });
    },
  });
};
