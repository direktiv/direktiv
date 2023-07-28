import { InstancesDetailSchema, InstancesDetailSchemaType } from "../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const getInstanceDetails = apiFactory({
  url: ({
    namespace,
    baseUrl,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) => `${baseUrl ?? ""}/api/namespaces/${namespace}/instances/${instanceId}`,
  method: "GET",
  schema: InstancesDetailSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<ReturnType<(typeof instanceKeys)["instanceDetail"]>>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceDetails = (
  { instanceId }: { instanceId: string },
  { stream }: { stream?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}`,
    enabled: !!stream,
    schema: InstancesDetailSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<InstancesDetailSchemaType>(
        instanceKeys.instanceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
        }),
        () => msg
      );
    },
  });

  return useQuery({
    queryKey: instanceKeys.instanceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    // disable queryFn when streaming is enabled (to avoid duplicate requests)
    enabled: !stream,
  });
};
