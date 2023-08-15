import {
  WorkflowVariableCreatedSchema,
  WorkflowVariableCreatedSchemaType,
  WorkflowVariableFormSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

const setVariable = apiFactory({
  url: ({
    namespace,
    path,
    name,
  }: {
    namespace: string;
    path: string;
    name: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=set-var&var=${name}`,
  method: "PUT",
  schema: WorkflowVariableCreatedSchema,
});

export const useSetVariable = ({
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
    payload,
  }: {
    name: string;
    path: string;
    payload: WorkflowVariableFormSchemaType;
  }) =>
    setVariable({
      apiKey: apiKey ?? undefined,
      payload,
      urlParams: {
        namespace,
        path,
        name,
      },
    });

  return useMutation({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.setQueryData<WorkflowVariableCreatedSchemaType>(
        treeKeys.workflowVariablesList(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        data
      );
      onSuccess?.(data);
      toast({
        title: t("api.tree.mutate.setVariable.success.title"),
        description: t("api.tree.mutate.setVariable.success.title", {
          variable: data.key,
          workflow: data.path,
        }),
      });
    },
  });
};
