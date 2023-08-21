import {
  PodsListSchema,
  PodsListSchemaType,
  PodsStreamingSchema,
  PodsStreamingSchemaType,
} from "../../../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { memo } from "react";
import { serviceKeys } from "../../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: PodsListSchemaType | undefined,
  streamingPayload: PodsStreamingSchemaType
) => {
  // only react top ADDED events
  if (streamingPayload.event !== "ADDED") return oldData;

  if (!oldData) {
    return {
      pods: [streamingPayload.pod],
    };
  }

  return {
    pods: oldData.pods.map((pod) =>
      pod.name === streamingPayload.pod.name ? streamingPayload.pod : pod
    ),
  };
};

export const getPods = apiFactory({
  url: ({
    baseUrl,
    namespace,
    service,
    revision,
  }: {
    baseUrl?: string;
    namespace: string;
    service: string;
    revision: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}/pods`,
  method: "GET",
  schema: PodsListSchema,
});

const fetchPods = async ({
  queryKey: [{ apiKey, namespace, service, revision }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicePods"]>>) =>
  getPods({
    apiKey,
    urlParams: { namespace, service, revision },
  });

export const usePodsStream = (
  {
    service,
    revision,
  }: {
    service: string;
    revision: string;
  },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}/pods`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: PodsStreamingSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<PodsListSchemaType>(
        serviceKeys.servicePods(namespace, {
          apiKey: apiKey ?? undefined,
          revision,
          service,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type ServicesStreamingSubscriberType = {
  service: string;
  revision: string;
  enabled?: boolean;
};

export const PodsSubscriber = memo(
  ({ service, revision, enabled }: ServicesStreamingSubscriberType) => {
    usePodsStream(
      {
        service,
        revision,
      },
      { enabled: enabled ?? true }
    );
    return null;
  }
);

PodsSubscriber.displayName = "PodsSubscriber";

export const usePods = ({
  service,
  revision,
}: {
  service: string;
  revision: string;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.servicePods(namespace, {
      apiKey: apiKey ?? undefined,
      revision,
      service,
    }),
    queryFn: fetchPods,
    enabled: !!namespace,
  });
};
