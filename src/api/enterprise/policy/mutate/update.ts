import { useMutation, useQueryClient } from "@tanstack/react-query";

import { PolicyCreatedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// TODO: remove the line below and delete the mock function
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const updatePolicy = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/policy`,
  method: "PUT",
  schema: PolicyCreatedSchema,
});

const updatePolicyMock = (params: {
  apiKey?: string;
  payload: string;
  urlParams: { namespace: string };
}): Promise<z.infer<typeof PolicyCreatedSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      console.warn("mocking save mutation with", params.payload);
      resolve(null);
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

  return useMutation({
    mutationFn: ({ policyContent }: { policyContent: string }) =>
      updatePolicyMock({
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
      const message = z
        .object({
          message: z.string(),
        })
        .safeParse(e);
      message.success ? onError?.(message.data.message) : onError?.(undefined);
    },
  });
};
