import { ServicesListSchema, ServicesListSchemaType } from "../schema/services";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";
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

/**
 * we only have one service endpoint, but we use this useServices hook
 * as a separate abstraction layer to derive the other more specific
 * hooks from.
 */
const useServices = <T>({
  filter,
}: {
  filter: (apiResponse: ServicesListSchemaType) => T;
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

export const useService = (service: string) =>
  useServices({
    filter: (apiResponse) =>
      apiResponse.data.find((serviceObj) => serviceObj.id === service),
  });

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
