import {
  WorkflowVariableCreatedSchema,
  WorkflowVariableCreatedSchemaType,
  WorkflowVariableFormSchemaType,
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

export const setVariable = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
    name,
  }: {
    baseUrl?: string;
    namespace: string;
    path: string;
    name: string;
  }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=set-var&var=${name}`,
  method: "PUT",
  schema: WorkflowVariableCreatedSchema,
});

export const useSetWorkflowVariable = ({
  onSuccess,
}: {
  onSuccess?: (data: WorkflowVariableCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({
    name,
    path,
    content,
    mimeType,
  }: WorkflowVariableFormSchemaType) =>
    setVariable({
      apiKey: apiKey ?? undefined,
      payload: content,
      urlParams: {
        namespace,
        path,
        name,
      },
      headers: {
        "Content-Type": mimeType,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries(
        treeKeys.workflowVariablesList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        })
      );
      queryClient.invalidateQueries(
        treeKeys.workflowVariableContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
          name: variables.name,
        })
      );
      onSuccess?.(data);
      toast({
        title: t("api.tree.mutate.setVariable.success.title"),
        description: t("api.tree.mutate.setVariable.success.title", {
          variable: data.key,
          workflow: data.path,
        }),
        variant: "success",
      });
    },
  });
};
