import {
  InstanceCancelPayloadType,
  InstanceCanceledResponseSchema,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";
import { useNamespace } from "~/util/store/namespace";

const updateInstance = apiFactory<InstanceCancelPayloadType>({
  url: ({
    baseUrl,
    namespace,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/${instanceId}`,
  method: "PATCH",
  schema: InstanceCanceledResponseSchema,
});

export const useCancelInstance = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = (instanceId: string) =>
    updateInstance({
      apiKey: apiKey ?? undefined,
      urlParams: {
        namespace,
        instanceId,
      },
      payload: {
        status: "cancelled",
      },
    });

  return useMutationWithPermissions({
    mutationFn,
  });
};
