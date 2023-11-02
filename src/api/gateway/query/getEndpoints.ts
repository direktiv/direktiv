import { EndpointListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getEndpoints = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/endpoints`,
  method: "GET",
  schema: EndpointListSchema,
});

const fetchEndpoints = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["endpoints"]>>) =>
  getEndpoints({
    apiKey,
    urlParams: { namespace },
  });

export const useEndpoints = ({
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
    queryKey: gatewayKeys.endpoints(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchEndpoints,
    enabled: !!namespace && enabled,
  });
};
