import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
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
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not update workflow 😢",
        variant: "error",
      });
    },
  });
};
