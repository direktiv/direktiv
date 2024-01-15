import { PodLogsSchema, PodLogsSchemaType } from "../schema/pods";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { memo } from "react";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/httpStreaming";

export const usePodLogsStream = (
  {
    namespace,
    service,
    pod,
  }: {
    namespace: string;
    service: string;
    pod: string;
  },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const queryClient = useQueryClient();

  useStreaming({
    url: `/api/v2/namespaces/${namespace}/services/${service}/pods/${pod}/logs`,
    apiKey: apiKey ?? undefined,
    schema: PodLogsSchema,
    enabled,
    onMessage: (data, isFirstMessage) => {
      queryClient.setQueryData<PodLogsSchemaType>(
        serviceKeys.podLogs({
          namespace,
          apiKey: apiKey ?? undefined,
          pod,
          service,
        }),
        (old) => {
          if (isFirstMessage) {
            return data;
          }
          return `${old ?? ""}${data}`;
        }
      );
    },
  });
};

type PodLogsSubscriberType = {
  service: string;
  pod: string;
  enabled?: boolean;
};

export const PodLogsSubscriber = memo(
  ({ service, pod, enabled }: PodLogsSubscriberType) => {
    const namespace = useNamespace();
    if (!namespace) {
      throw new Error("namespace is undefined");
    }
    usePodLogsStream(
      {
        namespace,
        service,
        pod,
      },
      { enabled: enabled ?? true }
    );
    return null;
  }
);

PodLogsSubscriber.displayName = "PodLogsSubscriber";

export const usePodLogs = ({
  service,
  pod,
}: {
  service: string;
  pod: string;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery<PodLogsSchemaType>({
    queryKey: serviceKeys.podLogs({
      namespace,
      apiKey: apiKey ?? undefined,
      service,
      pod,
    }),
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * will be added by a streaming subscriber. We don't have a non-streaming endpoint
     * for initial data. So the queryFn is missing on purpose and the enabled flag is set
     * to false.
     */
    enabled: false,
  });
};
