import { RoutesListSchema, RoutesListSchemaType } from "../schema";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
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

const useRoutesGeneric = <T>({
  filter,
  enabled = true,
}: {
  filter: (apiResponse: RoutesListSchemaType) => T;
  enabled?: boolean;
}) => {
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
    enabled: !!namespace && enabled,
    select: (data) => {
      // Ensure plugins properties are arrays
      const normalizedData = data.data.map((route) => ({
        ...route,
        spec: {
          ...route.spec,
          "x-direktiv-config": {
            ...route.spec["x-direktiv-config"],
            plugins: {
              ...route.spec["x-direktiv-config"].plugins,
              inbound: route.spec["x-direktiv-config"].plugins.inbound ?? [],
              outbound: route.spec["x-direktiv-config"].plugins.outbound ?? [],
              auth: route.spec["x-direktiv-config"].plugins.auth ?? [],
            },
          },
        },
      }));
      return filter({ data: normalizedData });
    },
  });
};

export const useRoutes = () =>
  useRoutesGeneric({
    filter: (apiResponse) => apiResponse,
  });

export const useRoute = ({
  routePath,
  enabled,
}: {
  routePath: string;
  enabled?: boolean;
}) =>
  useRoutesGeneric({
    filter: (apiResponse) =>
      apiResponse.data.find(
        (route) => route.file_path === forceLeadingSlash(routePath)
      ),
    enabled,
  });
