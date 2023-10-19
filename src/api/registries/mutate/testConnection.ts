import { RegistryTestConnectionSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";

export const testConnection = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) =>
    `${baseUrl ?? ""}/api/functions/registries/test`,
  method: "POST",
  schema: RegistryTestConnectionSchema,
});

export const useTestConnection = ({
  onSuccess,
  onError,
}: {
  onSuccess?: () => void;
  onError?: () => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutationWithPermissions({
    mutationFn: ({
      url,
      username,
      password,
    }: {
      url: string;
      username: string;
      password: string;
    }) =>
      testConnection({
        apiKey: apiKey ?? undefined,
        urlParams: {},
        payload: { url, username, password },
      }),
    onSuccess: () => {
      onSuccess?.();
    },
    onError: () => {
      onError?.();
    },
  });
};
