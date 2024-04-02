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
  url: ({ namespace, variableID }: { namespace: string; variableID: string }) =>
    `/api/v2/namespaces/${namespace}/variables/${variableID}`,
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
  const { t } = useTranslation();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ variable }: { variable: VarSchemaType }) =>
    deleteVar({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        variableID: variable.id,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(
        varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: t("api.variables.mutate.deleteVariable.success.title"),
        description: t(
          "api.variables.mutate.deleteVariable.success.description",
          { name: variables.variable.name }
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
