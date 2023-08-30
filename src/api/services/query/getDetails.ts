import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  RevisionStreamingSchema,
  RevisionStreamingSchemaType,
  RevisionsListSchema,
  RevisionsListSchemaType,
} from "../schema/revisions";

import { apiFactory } from "~/api/apiFactory";
import { memo } from "react";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const getServiceDetails = apiFactory({
  url: ({
    baseUrl,
    namespace,
    service,
  }: {
    baseUrl?: string;
    namespace: string;
    service: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/functions/namespaces/${namespace}/function/${service}`,
  method: "GET",
  schema: RevisionsListSchema,
});

const fetchServiceDetails = async ({
  queryKey: [{ apiKey, namespace, service }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["serviceDetail"]>>) =>
  getServiceDetails({
    apiKey,
    urlParams: { namespace, service },
  }).then((res) => ({
    // revisions must be sorted by creation date, to figure out the latest revision
    ...res,
    revisions: (res.revisions ?? []).sort((a, b) => {
      if (a.revision > b.revision) {
        return -1;
      }
      if (a.revision < b.revision) {
        return 1;
      }
      return 0;
    }),
  }));

const updateCache = (
  oldData: RevisionsListSchemaType | undefined,
  streamingPayload: RevisionStreamingSchemaType
) => {
  if (!oldData) {
    return undefined;
  }
  return {
    ...oldData,
    /**
     * Please note that we don't add new services to the react-query cache via streaming
     * because the streamingPayload.event === "ADDED" cannot be trusted. Instead, we 
     * update the cache manually. The drawback is this will not detect services added outside
     * of the current browser window.
     */
    revisions: (oldData.revisions ?? [])
      // swap the element that came in (if it already is in the cache)
      .map((rev) => ({
        /**
         * we need to merge the old revision, because we don't consume all fields
         * when streaming. The streaming payload has some minor inconsistencies
         * with the revision schema. However, the fields that we keep from the
         * cache are long living ones like the creation date and the name (which
         * acts like an id)
         */
        ...rev,
        ...(rev.name === streamingPayload.revision.name
          ? streamingPayload.revision
          : {}),
      }))
      // remove element if it was deleted
      .filter((rev) => {
        if (streamingPayload.event !== "DELETED") {
          return true;
        }
        return rev.name !== streamingPayload.revision.name;
      }),
  };
};

export const useServiceDetailsStream = (
  service: string,
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/functions/namespaces/${namespace}/function/${service}/revisions`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: RevisionStreamingSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<RevisionsListSchemaType>(
        serviceKeys.serviceDetail(namespace, {
          apiKey: apiKey ?? undefined,
          service,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type ServiceRevisionStreamingSubscriberType = {
  service: string;
  enabled?: boolean;
};

export const ServiceDetailsStreamingSubscriber = memo(
  ({ service, enabled }: ServiceRevisionStreamingSubscriberType) => {
    useServiceDetailsStream(service, { enabled: enabled ?? true });
    return null;
  }
);

ServiceDetailsStreamingSubscriber.displayName =
  "ServiceDetailsStreamingSubscriber";

export const useServiceDetails = ({ service }: { service: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.serviceDetail(namespace, {
      apiKey: apiKey ?? undefined,
      service,
    }),
    queryFn: fetchServiceDetails,
    enabled: !!service,
  });
};
