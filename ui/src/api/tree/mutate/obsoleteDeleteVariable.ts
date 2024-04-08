import {
  WorkflowVariableDeletedSchema,
  WorkflowVariableSchemaType,
} from "../schema/obsoleteWorkflowVariable";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const deleteWorkflowVariable = apiFactory({
  url: ({
    namespace,
    name,
    path,
  }: {
    namespace: string;
    name: string;
    path: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=delete-var&var=${name}`,
  method: "DELETE",
  schema: WorkflowVariableDeletedSchema,
});

export const useDeleteWorkflowVariable = ({
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
    variable,
    path,
  }: {
    variable: WorkflowVariableSchemaType;
    path: string;
  }) =>
    deleteWorkflowVariable({
      apiKey: apiKey ?? undefined,
      urlParams: {
        path,
        namespace,
        name: variable.name,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(
        treeKeys.workflowVariablesList(namespace, {
          path: variables.path,
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
