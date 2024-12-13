import { VarDeletedSchema, VarSchemaType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { varKeys } from "..";

const deleteVar = apiFactory({
  url: ({
    namespace,
    variableIDs,
  }: {
    namespace: string;
    variableIDs: string[];
  }) =>
    variableIDs.length === 0
      ? `/api/v2/namespaces/${namespace}/variables/${variableIDs[0]}`
      : `/api/v2/namespaces/${namespace}/variables?ids=${variableIDs.join(",")}`,
  method: "DELETE",
  schema: VarDeletedSchema,
});

export const useDeleteVar = ({
  onSuccess,
}: { onSuccess?: () => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (variables: VarSchemaType[]) =>
    deleteVar({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        variableIDs: variables.map((v) => v.id),
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (_, input) => {
      queryClient.invalidateQueries({
        queryKey: varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
          workflowPath:
            input[0]?.type === "workflow-variable"
              ? input[0].reference
              : undefined,
        }),
      });
      toast({
        title: t("api.variables.mutate.deleteVariable.success.title"),
        description: t(
          input.length === 1
            ? "api.variables.mutate.deleteVariable.success.description_one"
            : "api.variables.mutate.deleteVariable.success.description",
          { count: input.length }
        ),
        variant: "success",
      });
      onSuccess?.();
    },
    onError: () => {
      toast({
        title: t("api.generic.error"),
        description: t("api.variables.mutate.deleteVariable.error.description"),
        variant: "error",
      });
    },
  });
};
