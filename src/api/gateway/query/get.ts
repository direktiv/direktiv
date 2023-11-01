import { GatewayListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getGatewayList = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/endpoints`,
  method: "GET",
  schema: GatewayListSchema,
});

const fetchGatewayList = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["servicesList"]>>) =>
  getGatewayList({
    apiKey,
    urlParams: { namespace },
  });

export const useGatewayList = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.servicesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchGatewayList,
    enabled: !!namespace,
  });
};
