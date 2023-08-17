import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  ServiceStreamingSchema,
  ServiceStreamingSchemaType,
  ServicesListSchema,
  ServicesListSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { memo } from "react";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export const getServices = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/functions/namespaces/${namespace}`,
  method: "GET",
  schema: ServicesListSchema,
});

const fetchServices = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicesList"]>>) =>
  getServices({
    apiKey,
    urlParams: { namespace },
  }).then((res) => ({
    // TODO: this should be changed in the backend
    // reverse the order of functions (newer first)
    ...res,
    functions: [...(res.functions ?? []).reverse()],
  }));

const updateCache = (
  oldData: ServicesListSchemaType | undefined,
  streamingPayload: ServiceStreamingSchemaType
) => {
  if (!oldData) {
    return undefined;
  }

  return {
    ...oldData,
    /**
     * please note that that streaming the services will never add new services
     * because the streamingPayload.event === "ADDED" can not be trusted. This
     * would also just be usefull to recognize new services that are note created
     * by the window that is currently open. The user might not except that anyways
     */
    functions: oldData.functions
      // swap the element that came in (if it already is in the cache)
      .map((func) => {
        if (func.serviceName === streamingPayload.function.serviceName) {
          return streamingPayload.function;
        }
        return func;
      })
      // remove element if it was deleted
      .filter((func) => {
        if (streamingPayload.event !== "DELETED") {
          return true;
        }
        return func.serviceName !== streamingPayload.function.serviceName;
      }),
  };
};

export const useServicesStream = ({
  enabled = true,
}: { enabled?: boolean } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/functions/namespaces/${namespace}`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: ServiceStreamingSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<ServicesListSchemaType>(
        serviceKeys.servicesList(namespace, {
          apiKey: apiKey ?? undefined,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type ServicesStreamingSubscriberType = {
  enabled?: boolean;
};

export const InstanceStreamingSubscriber = memo(
  ({ enabled }: ServicesStreamingSubscriberType) => {
    useServicesStream({ enabled: enabled ?? true });
    return null;
  }
);

InstanceStreamingSubscriber.displayName = "InstanceStreamingSubscriber";

export const useServices = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: serviceKeys.servicesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchServices,
    enabled: !!namespace,
  });
};
