import { WorkflowStartedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "~/util/store/namespace";

export const runWorkflow = apiFactory({
  url: ({
    baseUrl,
    namespace,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path?: string;
  }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=execute&ref=latest`,
  method: "POST",
  schema: WorkflowStartedSchema,
});

type ResolvedRunWorkflow = Awaited<ReturnType<typeof runWorkflow>>;

export const useRunWorkflow = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedRunWorkflow) => void } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path, payload }: { path: string; payload: string }) =>
      runWorkflow({
        apiKey: apiKey ?? undefined,
        payload,
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess: (data) => {
      onSuccess?.(data);
    },
  });
};
