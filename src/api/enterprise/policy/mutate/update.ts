import { PolicyCreatedSchema } from "../schema";
import { getMessageFromApiError } from "~/api/errorHandling";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

// const updatePolicy = apiFactory({
//   url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
//     `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/policy`,
//   method: "PUT",
//   schema: PolicyCreatedSchema,
// });

const updatePolicy = (_params: {
  apiKey?: string;
  payload: string;
  urlParams: { namespace: string };
}): Promise<z.infer<typeof PolicyCreatedSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve({});
    }, 1000);
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
