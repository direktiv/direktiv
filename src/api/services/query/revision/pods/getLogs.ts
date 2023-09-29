import { PodLogsSchema, PodLogsSchemaType } from "../../../schema/pods";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { memo } from "react";
import { serviceKeys } from "../../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const usePodLogsStream = (
  {
    name,
    namespace,
  }: {
    name: string;
    namespace: string;
  },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const queryClient = useQueryClient();

  return useStreaming({
    url: `/api/functions/namespaces/${namespace}/logs/pod/${name}`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: PodLogsSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<PodLogsSchemaType>(
        serviceKeys.podLogs({
          namespace,
          name,
          apiKey: apiKey ?? undefined,
        }),
        // the sreaming endpoint just returns the new cache value
        () => msg
      );
    },
  });
};

type PodLogsSubscriberType = {
  name: string;
  enabled?: boolean;
};

export const PodLogsSubscriber = memo(
  ({ name, enabled }: PodLogsSubscriberType) => {
    const namespace = useNamespace();

    if (!namespace) {
      throw new Error("namespace is undefined");
    }

    usePodLogsStream(
      {
        name,
        namespace,
      },
      { enabled: enabled ?? true }
    );
    return null;
  }
);

PodLogsSubscriber.displayName = "PodLogsSubscriber";

export const usePodLogs = ({ name }: { name: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery<PodLogsSchemaType>({
    queryKey: serviceKeys.podLogs({
      apiKey: apiKey ?? undefined,
      name,
      namespace,
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
