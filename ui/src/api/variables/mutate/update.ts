import {
  VarCreatedUpdatedSchema,
  VarFormCreateEditSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { varKeys } from "..";

const updateVar = apiFactory<VarFormCreateEditSchemaType>({
  url: ({
    baseUrl,
    namespace,
    id,
  }: {
    baseUrl?: string;
    namespace: string;
    id: string;
  }) => `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/variables/${id}`,
  method: "PATCH",
  schema: VarCreatedUpdatedSchema,
});

export const useUpdateVar = ({
  onSuccess,
}: {
  onSuccess?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({
    id,
    ...payload
  }: { id: string } & VarFormCreateEditSchemaType) =>
    updateVar({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        id,
        namespace,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: varKeys.varDetails(namespace, {
          apiKey: apiKey ?? undefined,
          id: variables.id,
        }),
      });
      // the list also needs to be invalidated because the variable's name could have changed
      queryClient.invalidateQueries({
        queryKey: varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
          workflowPath:
            data.data.type === "workflow-variable"
              ? data.data.reference
              : undefined,
        }),
      });
      toast({
        title: t("api.variables.mutate.updateVariable.success.title"),
        description: t(
          "api.variables.mutate.updateVariable.success.description",
          { name: data.data.name }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.variables.mutate.updateVariable.error.description"),
        variant: "error",
      });
    },
  });
};
