import {
  InstanceDetailsResponseSchemaType,
  InstanceDetailsSchema,
} from "../../schema";

import { instanceKeys } from "../..";
import { memo } from "react";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useStreaming } from "~/api/streaming";

type InstanceStreamingParams = {
  instanceId: string;
  enabled?: boolean;
};

export const useInstanceDetailsStream = ({
  instanceId,
  enabled,
}: InstanceStreamingParams) => {
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

export const InstanceStreamingSubscriber = memo(
  (params: InstanceStreamingParams) => {
    useInstanceDetailsStream({ ...params });
    return null;
  }
);

InstanceStreamingSubscriber.displayName = "InstanceStreamingSubscriber";
