import { OpenapiSpecificationSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getInfo = apiFactory({
  url: ({
    baseUrl,
    namespace,
    expand,
  }: {
    baseUrl?: string;
    namespace: string;
    expand?: boolean;
  }) => {
    const url = `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/info`;

    if (!expand) {
      return url;
    }

    const serverParams = `${window.location.origin}/ns/${namespace}`;
    const queryParams = new URLSearchParams();
    queryParams.append("expand", "true");
    queryParams.append("server", serverParams);

    return `${url}?${queryParams.toString()}`;
  },

  method: "GET",
  schema: OpenapiSpecificationSchema,
});

const fetchInfo = async ({
  queryKey: [{ apiKey, namespace, expand }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["info"]>>) =>
  getInfo({
    apiKey,
    urlParams: { namespace, expand },
  });

export const useInfo = ({
  expand,
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
    }),
    queryFn: fetchInfo,
    enabled: !!namespace,
  });
};
