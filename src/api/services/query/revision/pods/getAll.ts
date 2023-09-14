import {
  PodsListSchema,
  PodsListSchemaType,
  PodsStreamingSchema,
  PodsStreamingSchemaType,
} from "../../../schema/pods";
import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
import { memo } from "react";
import { serviceKeys } from "../../..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: PodsListSchemaType | undefined,
  streamingPayload: PodsStreamingSchemaType
) => {
  // only react to ADDED events
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
    workflow,
    version,
  }: {
    baseUrl?: string;
    namespace: string;
    service: string;
    revision: string;
    workflow?: string;
    version?: string;
  }) => {
    let url;

    if (workflow && version) {
      url = `${
        baseUrl ?? ""
      }/api/functions/namespaces/${namespace}/tree${forceLeadingSlash(
        workflow
      )}?op=pods&svn=${service}&rev=${revision}&version=${version}`;
    } else {
      url = `${
        baseUrl ?? ""
      }/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}/pods`;
    }

    return url;
  },
  method: "GET",
  schema: PodsListSchema,
});

const fetchPods = async ({
  queryKey: [{ apiKey, namespace, service, revision, workflow, version }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicePods"]>>) =>
  getPods({
    apiKey,
    urlParams: { namespace, service, revision, workflow, version },
  });

export const usePodsStream = (
  {
    service,
    revision,
    workflow,
    version,
  }: {
    service: string;
    revision: string;
    workflow?: string;
    version?: string;
  },
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
        )}?op=pods&svn=${service}&rev=${revision}&version=${version}`
      : `/api/functions/namespaces/${namespace}/function/${service}/revisions/${revision}/pods`;

  return useStreaming({
    url,
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
  workflow?: string;
  version?: string;
  enabled?: boolean;
};

export const PodsSubscriber = memo(
  ({
    service,
    revision,
    workflow,
    version,
    enabled,
  }: ServicesStreamingSubscriberType) => {
    usePodsStream(
      {
        service,
        revision,
        workflow,
        version,
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
  workflow,
  version,
}: {
  service: string;
  revision: string;
  workflow?: string;
  version?: string;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: serviceKeys.servicePods(namespace, {
      apiKey: apiKey ?? undefined,
      revision,
      service,
      workflow,
      version,
    }),
    queryFn: fetchPods,
    enabled: !!namespace,
  });
};
