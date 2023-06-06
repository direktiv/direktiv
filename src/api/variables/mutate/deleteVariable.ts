import { VarDeletedSchema, VarSchemaType } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { varKeys } from "..";

const deleteVar = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "DELETE",
  schema: VarDeletedSchema,
});

export const useDeleteVar = ({
  onSuccess,
}: {
  onSuccess?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ variable }: { variable: VarSchemaType }) =>
    deleteVar({
      apiKey: apiKey ?? undefined,
      payload: undefined,
      urlParams: {
        namespace: namespace,
        name: variable.name,
      },
      headers: undefined,
    });

  return useMutation({
    mutationFn,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: "Variable deleted",
        description: `Variable ${variables.variable.name} was deleted.`,
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "Could not delete variable 😢",
        variant: "error",
      });
    },
  });
};
