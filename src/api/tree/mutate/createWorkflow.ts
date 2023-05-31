import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "~/api/utils";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";

export const createWorkflow = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
    name,
  }: {
    baseUrl?: string;
    namespace: string;
    path?: string;
    name: string;
  }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}/${name}?op=create-workflow`,
  method: "PUT",
  schema: WorkflowCreatedSchema,
});

type ResolvedCreateWorkflow = Awaited<ReturnType<typeof createWorkflow>>;

export const useCreateWorkflow = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateWorkflow) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({
      path,
      name,
      fileContent,
    }: {
      path?: string;
      name: string;
      fileContent: string;
    }) =>
      createWorkflow({
        apiKey: apiKey ?? undefined,
        payload: fileContent,
        headers: undefined,
        urlParams: {
          namespace: namespace,
          path,
          name,
        },
      }),
    onSuccess(data, variables) {
      toast({
        title: "Workflow created",
        description: `Workflow ${variables.name} was created in ${variables.path}`,
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create workflow 😢",
        variant: "error",
      });
    },
  });
};
