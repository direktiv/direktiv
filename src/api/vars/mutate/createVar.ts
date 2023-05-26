import { VarCreatedSchema, VarCreatedSchemaType } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { varKeys } from "..";

const createVar = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "PUT",
  schema: VarCreatedSchema,
});

export const useCreateVar = ({
  onSuccess,
}: {
  onSuccess?: (data: VarCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ name, content }: { name: string; content: string }) =>
    createVar({
      apiKey: apiKey ?? undefined,
      payload: content,
      urlParams: {
        namespace: namespace,
        name,
      },
    });

  return useMutation({
    mutationFn,
    onSuccess: (data) => {
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: "Variable created",
        description: `Variable ${data.key} was created.`,
        variant: "success",
      });
      onSuccess?.(data);
    },
  });
};
