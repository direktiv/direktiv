import { PolicyCreatedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { getMessageFromApiError } from "~/api/errorHandling";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";

const updatePolicy = apiFactory<string>({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/policy`,
  method: "PUT",
  schema: PolicyCreatedSchema,
});

export const useUpdatePolicy = ({
  onSuccess,
  onError,
}: {
  onSuccess?: () => void;
  onError?: (e: string | undefined) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({ policyContent }: { policyContent: string }) =>
      updatePolicy({
        apiKey: apiKey ?? undefined,
        payload: policyContent,
        urlParams: {
          namespace,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries(
        policyKeys.get(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      onSuccess?.();
    },
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};
