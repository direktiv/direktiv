import {
  InstanceDetailsResponseSchema,
  InstanceDetailsResponseSchemaType,
  InstanceDetailsSchema,
} from "../schema";
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
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/instances/${instanceId}`,
  method: "GET",
  schema: InstanceDetailsResponseSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace, instanceId }],
}: QueryFunctionContext<
  ReturnType<(typeof instanceKeys)["instancesDetails"]>
>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace, instanceId },
  });

export const useInstanceDetailsStream = ({
  instanceId,
  enabled,
}: {
  instanceId: string;
  enabled: boolean;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/v2/namespaces/${namespace}/instances/${instanceId}/subscribe`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: InstanceDetailsSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<InstanceDetailsResponseSchemaType>(
        instanceKeys.instancesDetails(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
        }),
        () => ({
          data: msg,
        })
      );
    },
  });
};

type InstanceStreamingSubscriberType = {
  instanceId: string;
  enabled?: boolean;
};

export const InstanceStreamingSubscriber = memo(
  ({ instanceId, enabled = true }: InstanceStreamingSubscriberType) => {
    useInstanceDetailsStream({ instanceId, enabled });
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
    queryKey: instanceKeys.instancesDetails(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
    select: (data) => data.data,
  });
};
