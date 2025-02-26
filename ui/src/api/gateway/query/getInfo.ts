import { OpenapiSpecificationSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getInfo = apiFactory({
  url: ({
    baseUrl,
    namespace,
    expand,
    server,
  }: {
    baseUrl?: string;
    namespace: string;
    expand?: boolean;
    server?: string;
  }) => {
    const url = `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/info`;

    const queryParams = new URLSearchParams();
    if (expand) queryParams.append("expand", "true");
    if (server) queryParams.append("server", server);

    return queryParams.toString() ? `${url}?${queryParams.toString()}` : url;
  },

  method: "GET",
  schema: OpenapiSpecificationSchema,
});

const fetchInfo = async ({
  queryKey: [{ apiKey, namespace, expand, server }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["info"]>>) =>
  getInfo({
    apiKey,
    urlParams: { namespace, expand, server },
  });

export const useInfo = ({
  expand,
  server,
}: { expand?: boolean; server?: string } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.info(namespace, {
      apiKey: apiKey ?? undefined,
      expand,
      server: server ?? `${window.location.origin}/ns/${namespace}`,
    }),
    queryFn: fetchInfo,
    enabled: !!namespace,
  });
};
