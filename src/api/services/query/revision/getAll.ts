import {
  ServiceRevisionDetailSchemaType,
  ServiceRevisionDetailStreamingSchema,
  ServiceRevisionDetailStreamingSchemaType,
} from "../../schema";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { memo } from "react";
import { serviceKeys } from "../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: ServiceRevisionDetailSchemaType | undefined,
  streamingPayload: ServiceRevisionDetailStreamingSchemaType
) => {
  if (streamingPayload.event === "ADDED") {
    return streamingPayload.revision;
  }
};

export const useServiceRevisionStream = (
  { service, revision }: { service: string; revision: string },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: ServiceRevisionDetailStreamingSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<ServiceRevisionDetailSchemaType>(
        serviceKeys.serviceRevisionDetail(namespace, {
          apiKey: apiKey ?? undefined,
          service,
          revision,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type ServiceRevisionStreamingSubscriberType = {
  service: string;
  revision: string;
  enabled?: boolean;
};

export const ServiceRevisionStreamingSubscriber = memo(
  ({ service, revision, enabled }: ServiceRevisionStreamingSubscriberType) => {
    useServiceRevisionStream(
      { service, revision },
      { enabled: enabled ?? true }
    );
    return null;
  }
);

ServiceRevisionStreamingSubscriber.displayName =
  "ServiceRevisionStreamingSubscriber";

export const useServiceRevision = ({
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
    queryKey: serviceKeys.serviceRevisionDetail(namespace, {
      apiKey: apiKey ?? undefined,
      service,
      revision,
    }),
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * will be added by a streaming subscriber. We don't have any non streaming endpoint
     * to initial data. So the queryFn is missing on purpose and the enabled flag is set
     * to false.
     */
    enabled: false,
  });
};
