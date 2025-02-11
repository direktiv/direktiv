import { apiFactory } from "~/api/apiFactory";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

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
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/?path=${path}/&wait=true`,
  method: "POST",
  schema: z.any(), // to do change this to a good schema!
});

export const useCreateInstanceWithOutput = ({
  onSuccess,
  onError,
}: {
  onSuccess?: (namespace: string, data: any) => void;
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
