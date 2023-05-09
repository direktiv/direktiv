import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const createRevision = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=save-workflow&ref=latest`,
  method: "POST",
  schema: WorkflowCreatedSchema,
});

export const useCreateRevision = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path }: { path: string }) =>
      createRevision({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          namespace: namespace,
          path,
        },
      }),
    onSuccess(data) {
      toast({
        title: "Revision created",
        description: `Revision ${data.revision.name} was created`,
        variant: "success",
      });
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
