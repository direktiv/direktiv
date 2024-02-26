import { WorkflowStartedSchema } from "../schema/node";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
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
    )}?op=execute`,
  method: "POST",
  schema: WorkflowStartedSchema,
});

type ResolvedRunWorkflow = Awaited<ReturnType<typeof runWorkflow>>;

export const useRunWorkflow = ({
  onSuccess,
  onError,
}: {
  onSuccess?: (data: ResolvedRunWorkflow) => void;
  onError?: (error?: string) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
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
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};
