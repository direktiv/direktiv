import { GatewayListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getEndpointList = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/endpoints`,
  method: "GET",
  schema: GatewayListSchema,
});

const fetchEndpointList = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["endpointList"]>>) =>
  getEndpointList({
    apiKey,
    urlParams: { namespace },
  });

export const useEndpointList = ({
  enabled = true,
}: {
  enabled?: boolean;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.endpointList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchEndpointList,
    enabled: !!namespace && enabled,
  });
};
