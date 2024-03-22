import { InstanceCancelSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";

export const cancelInstance = apiFactory({
  url: ({
    baseUrl,
    namespace,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/namespaces/${namespace}/instances/${instanceId}/cancel`,
  method: "POST",
  schema: InstanceCancelSchema,
});

export const useCancelInstance = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (instanceId: string) =>
    cancelInstance({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        instanceId,
      },
    });

  return useMutationWithPermissions({
    mutationFn,
  });
};
