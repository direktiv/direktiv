import { OpenApiBaseFileSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getInfo = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/info`,
  method: "GET",
  schema: OpenApiBaseFileSchema,
});

const fetchInfo = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["info"]>>) =>
  getInfo({
    apiKey,
    urlParams: { namespace },
  });

export const useInfo = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.info(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchInfo,
    enabled: !!namespace,
  });
};
