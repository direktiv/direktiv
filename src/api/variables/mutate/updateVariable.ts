import {
  VarFormSchemaType,
  VarUpdatedSchema,
  VarUpdatedSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { varKeys } from "..";

const updateVar = apiFactory({
  url: ({ namespace, name }: { namespace: string; name: string }) =>
    `/api/namespaces/${namespace}/vars/${name}`,
  method: "PUT",
  schema: VarUpdatedSchema,
});

// This mutation has two use cases: creating a variable and updating
// a variable. Both use the same endpoint and verb in the backend API.
export const useUpdateVar = ({
  onSuccess,
}: {
  onSuccess?: (data: VarUpdatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ name, content, mimeType }: VarFormSchemaType) =>
    updateVar({
      apiKey: apiKey ?? undefined,
      payload: content,
      urlParams: {
        namespace: namespace,
        name,
      },
      headers: {
        "content-type": mimeType,
      },
    });

  return useMutation({
    mutationFn,
    onSuccess: (data, variables) => {
      // Two cache invalidations are needed due to the current API,
      // which uses the same endpoint for creating and editing.
      // varContent needs a refresh after editing, varList needs a
      // refresh after creating (the variable's content is not
      // included in the list)
      queryClient.invalidateQueries(
        varKeys.varContent(namespace, {
          apiKey: apiKey ?? undefined,
          name: variables.name,
        })
      );
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: "Variable saved",
        description: `Variable ${data.key} was saved.`,
        variant: "success",
      });
      onSuccess?.(data);
    },
  });
};
