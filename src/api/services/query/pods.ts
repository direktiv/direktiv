import { PodsListSchema } from "../schema/pods";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { serviceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getPods = apiFactory({
  url: ({
    baseUrl,
    namespace,
    serviceId,
  }: {
    baseUrl?: string;
    namespace: string;
    serviceId: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/v2/namespaces/${namespace}/services/${serviceId}/pods`,
  method: "GET",
  schema: PodsListSchema,
});

const fetchPods = async ({
  queryKey: [{ apiKey, namespace, serviceId }],
}: QueryFunctionContext<ReturnType<(typeof serviceKeys)["servicePods"]>>) =>
  getPods({
    apiKey,
    urlParams: { namespace, serviceId },
  });

export const usePods = (serviceId: string) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: serviceKeys.servicePods(namespace, {
      apiKey: apiKey ?? undefined,
      serviceId,
    }),
    queryFn: fetchPods,
    enabled: !!namespace,
  });
};
