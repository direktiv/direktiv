import { InstancesDetailSchema, InstancesDetailSchemaType } from "../schema";
import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { instanceKeys } from "..";
import { memo } from "react";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
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

export const useInstanceDetailsStream = (
  { instanceId }: { instanceId: string },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}`,
    apiKey: apiKey ?? undefined,
    enabled,
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
};

type InstanceStreamingSubscriberType = {
  instanceId: string;
  enabled?: boolean;
};

export const InstanceStreamingSubscriber = memo(
  ({ instanceId, enabled }: InstanceStreamingSubscriberType) => {
    useInstanceDetailsStream({ instanceId }, { enabled: enabled ?? true });
    return null;
  }
);

InstanceStreamingSubscriber.displayName = "InstanceStreamingSubscriber";

export const useInstanceDetails = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: instanceKeys.instanceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
  });
};
