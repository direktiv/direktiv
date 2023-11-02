import { PluginsListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getPlugins = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/plugins`,
  method: "GET",
  schema: PluginsListSchema,
});

const fetchPlugins = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["plugins"]>>) =>
  getPlugins({
    apiKey,
    urlParams: { namespace },
  });

export const usePlugins = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.plugins(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchPlugins,
    enabled: !!namespace,
    staleTime: Infinity,
  });
};
