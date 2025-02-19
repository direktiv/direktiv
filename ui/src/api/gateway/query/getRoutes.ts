import {
  MethodsKeys,
  OperationSchema,
  RoutesListSchema,
  RoutesListSchemaType,
  routeMethods,
} from "../schema";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { z } from "zod";

type OperationType = z.infer<typeof OperationSchema>;

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
      const normalizedData = data.data.map((route) => {
        const { spec } = route;
        const { "x-direktiv-config": config, ...rest } = spec;

        function isRouteMethod(
          key: string
        ): key is (typeof routeMethods)[number] {
          return routeMethods.includes(key as (typeof routeMethods)[number]);
        }

        const methods = Object.keys(rest).filter(isRouteMethod);

        const typedMethods: Partial<Record<MethodsKeys, OperationType>> =
          methods.reduce(
            (acc, method) => {
              acc[method as MethodsKeys] = rest[method];
              return acc;
            },
            {} as Partial<Record<MethodsKeys, OperationType>>
          );

        return {
          ...route,
          spec: {
            "x-direktiv-api": spec["x-direktiv-api"],
            "x-direktiv-config": {
              ...config,
              plugins: {
                ...config.plugins,
                inbound: config.plugins.inbound ?? [],
                outbound: config.plugins.outbound ?? [],
                auth: config.plugins.auth ?? [],
              },
            },
            ...Object.fromEntries(
              Object.keys(typedMethods).map((method) => [
                method,
                typedMethods[method as MethodsKeys] ?? {
                  description: "",
                  responses: {},
                },
              ])
            ),
          },
        };
      });
      return filter({ data: normalizedData });
    },
  });
};

export const useRoutes = () => {
  const namespace = useNamespace();
  if (!namespace) {
    console.warn(
      "useRoutes: namespace is undefined, query will not be enabled."
    );
  }
  return useRoutesGeneric({
    filter: (apiResponse) => apiResponse,
    enabled: !!namespace,
  });
};

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
