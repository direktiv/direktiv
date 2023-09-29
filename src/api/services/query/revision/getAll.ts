import {
  RevisionDetailSchemaType,
  RevisionDetailStreamingSchema,
  RevisionDetailStreamingSchemaType,
} from "../../schema/revisions";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { forceLeadingSlash } from "~/api/tree/utils";
import { memo } from "react";
import { serviceKeys } from "../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: RevisionDetailSchemaType | undefined,
  streamingPayload: RevisionDetailStreamingSchemaType
) => {
  if (streamingPayload.event === "ADDED") {
    return streamingPayload.revision;
  }
};

export const useServiceRevisionStream = (
  {
    service,
    revision,
    workflow,
    version,
  }: { service: string; revision: string; workflow?: string; version?: string },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const url =
    workflow && version
      ? `/api/functions/namespaces/${namespace}/tree${forceLeadingSlash(
          workflow
        )}?op=function-revision&svn=${service}&rev=${revision}&version=${version}`
      : `/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}`;

  return useStreaming({
    url,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: RevisionDetailStreamingSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<RevisionDetailSchemaType>(
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
  workflow?: string;
  version?: string;
  enabled?: boolean;
};

export const ServiceRevisionStreamingSubscriber = memo(
  ({
    service,
    revision,
    workflow,
    version,
    enabled,
  }: ServiceRevisionStreamingSubscriberType) => {
    useServiceRevisionStream(
      { service, revision, workflow, version },
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

  return useQuery<RevisionDetailSchemaType>({
    queryKey: serviceKeys.serviceRevisionDetail(namespace, {
      apiKey: apiKey ?? undefined,
      service,
      revision,
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
