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
    variableID,
    variableIDs,
  }: {
    namespace: string;
    variableID?: string;
    variableIDs?: string[];
  }) =>
    variableIDs
      ? `/api/v2/namespaces/${namespace}/variables?ids=${variableIDs.join(",")}`
      : `/api/v2/namespaces/${namespace}/variables/${variableID}`,
  method: "DELETE",
  schema: VarDeletedSchema,
});

type DeleteVarInput =
  | { variable: VarSchemaType }
  | { variables: VarSchemaType[] };

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

  const mutationFn = (input: DeleteVarInput) => {
    if ("variable" in input) {
      return deleteVar({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          variableID: input.variable.id,
        },
      });
    } else {
      if (!input.variables?.length) {
        throw new Error("No variables provided for deletion");
      }
      return deleteVar({
        apiKey: apiKey ?? undefined,
        urlParams: {
          namespace,
          variableIDs: input.variables.map((v) => v.id),
        },
      });
    }
  };

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (_, input) => {
      const variables =
        "variable" in input ? [input.variable] : input.variables;
      queryClient.invalidateQueries({
        queryKey: varKeys.varList(namespace, {
          apiKey: apiKey ?? undefined,
          workflowPath:
            variables[0]?.type === "workflow-variable"
              ? variables[0].reference
              : undefined,
        }),
      });
      toast({
        title: t("api.variables.mutate.deleteVariable.success.title"),
        description: t(
          "api.variables.mutate.deleteVariable.success.description",
          {
            name:
              variables.length > 1
                ? `${variables.length} variables`
                : variables[0]?.name || "",
          }
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
