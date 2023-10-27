import { ServicesListSchema, ServicesListSchemaType } from "../schema/services";
import { forceLeadingSlash, removeTrailingSlash } from "~/api/tree/utils";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { memo } from "react";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getServices = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/services`,
  method: "GET",
  schema: ServicesListSchema,
});

const fetchServices = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicesList"]>>) =>
  getServices({
    apiKey,
    urlParams: { namespace },
  });

// const updateCache = (
//   oldData: ServicesListSchemaType | undefined,
//   streamingPayload: ServiceStreamingSchemaType
// ) => {
//   if (!oldData) {
//     return undefined;
//   }

//   return {
//     ...oldData,
//     /**
//      * Please note that we don't add new services to the react-query cache via streaming
//      * because the streamingPayload.event === "ADDED" cannot be trusted. Instead, we
//      * update the cache manually. The drawback is this will not detect services added outside
//      * of the current browser window.
//      */
//     functions: oldData.functions
//       // swap the element that came in (if it already is in the cache)
//       .map((func) => {
//         if (func.serviceName === streamingPayload.function.serviceName) {
//           return streamingPayload.function;
//         }
//         return func;
//       })
//       // remove element if it was deleted
//       .filter((func) => {
//         if (streamingPayload.event !== "DELETED") {
//           return true;
//         }
//         return func.serviceName !== streamingPayload.function.serviceName;
//       }),
//   };
// };

// export const useServicesStream = ({
//   enabled = true,
//   workflow,
// }: { enabled?: boolean; workflow?: string } = {}) => {
//   const apiKey = useApiKey();
//   const namespace = useNamespace();
//   const queryClient = useQueryClient();

//   if (!namespace) {
//     throw new Error("namespace is undefined");
//   }

//   return useStreaming({
//     url: workflow
//       ? `/api/functions/namespaces/${namespace}/tree${forceLeadingSlash(
//           workflow
//         )}?op=services`
//       : `/api/functions/namespaces/${namespace}`,
//     apiKey: apiKey ?? undefined,
//     enabled,
//     schema: ServiceStreamingSchema,
//     onMessage: (msg) => {
//       queryClient.setQueryData<ServicesListSchemaType>(
//         serviceKeys.servicesList(namespace, {
//           apiKey: apiKey ?? undefined,
//           workflow: workflow ?? undefined,
//         }),
//         (oldData) => updateCache(oldData, msg)
//       );
//     },
//   });
// };

// type ServicesStreamingSubscriberType = {
//   enabled?: boolean;
//   workflow?: string;
// };

export const ServicesStreamingSubscriber = memo(
  () => null
  // ({ enabled, workflow }: ServicesStreamingSubscriberType) => {
  //   useServicesStream({
  //     enabled: enabled ?? true,
  //     workflow: workflow ?? undefined,
  //   });
  //   return null;
  // }
);

ServicesStreamingSubscriber.displayName = "ServicesStreamingSubscriber";

// TODO: rename this file or split it up
export const useService = (serviceId: string) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: serviceKeys.servicesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    select: (services) =>
      services?.data.find((service) => service.id === serviceId),
    queryFn: fetchServices,
    enabled: !!namespace,
  });
};

// TODO: don't export this hook
export const useServices = ({
  filter,
}: {
  filter: (apiResponse: ServicesListSchemaType) => ServicesListSchemaType;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: serviceKeys.servicesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchServices,
    enabled: !!namespace,
    select: (data) => filter(data),
  });
};

export const useWorkflowServices = (workflow: string) =>
  useServices({
    filter: (apiResponse) => ({
      data: apiResponse.data.filter(
        (service) =>
          service.type === "workflow-service" &&
          service.filePath === forceLeadingSlash(workflow)
      ),
    }),
  });

export const useNamespaceServices = () =>
  useServices({
    filter: (apiResponse) => ({
      data: apiResponse.data.filter(
        (service) => service.type === "namespace-service"
      ),
    }),
  });
