import { QueryFunctionContext } from "@tanstack/react-query";
import { RoutesListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getRoutes = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/routes`,
  method: "GET",
  schema: RoutesListSchema,
});

const fetchRoutes = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["routes"]>>) =>
  getRoutes({
    apiKey,
    urlParams: { namespace },
  });

export const useRoutes = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.routes(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchRoutes,
    enabled: !!namespace,
  });
};
