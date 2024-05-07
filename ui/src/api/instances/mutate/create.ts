import { InstanceCreatedResponseSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";

export const createInstance = apiFactory<string>({
  url: ({
    baseUrl,
    namespace,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path: string;
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/?path=${path}`,
  method: "POST",
  schema: InstanceCreatedResponseSchema,
});

type ResolvedRunWorkflow = Awaited<ReturnType<typeof createInstance>>;

export const useCreateInstance = ({
  onSuccess,
  onError,
}: {
  onSuccess?: (namespace: string, data: ResolvedRunWorkflow) => void;
  onError?: (error?: string) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ path, payload }: { path: string; payload: string }) =>
      createInstance({
        apiKey: apiKey ?? undefined,
        payload,
        urlParams: {
          namespace,
          path,
        },
      }),
    onSuccess: (data) => {
      onSuccess?.(namespace, data);
    },
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};
