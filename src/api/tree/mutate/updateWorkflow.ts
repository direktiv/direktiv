import { TreeListSchemaType, WorkflowCreatedSchema } from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const updateWorkflow = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=update-workflow`,
  method: "POST",
  schema: WorkflowCreatedSchema,
});

export const useUpdateWorkflow = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({
      path,
      fileContent,
    }: {
      path: string;
      fileContent: string;
    }) =>
      updateWorkflow({
        apiKey: apiKey ?? undefined,
        params: fileContent,
        pathParams: {
          namespace: namespace,
          path,
        },
      }),
    onSuccess: (data, variables) => {
      queryClient.setQueryData<TreeListSchemaType>(
        treeKeys.nodeContent(namespace, {
          apiKey: apiKey ?? undefined,
          path: variables.path,
        }),
        () => data
      );
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not update workflow 😢",
        variant: "error",
      });
    },
  });
};
