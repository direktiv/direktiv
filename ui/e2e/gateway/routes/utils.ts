import { NewRouteSchemaType, routeMethods } from "~/api/gateway/schema";

import { getRoutes } from "~/api/gateway/query/getRoutes";
import { headers } from "e2e/utils/testutils";

type CreateRouteFileParams = {
  path?: string;
  targetType?: string;
  targetConfigurationStatus?: string;
  enabledMethods?: (typeof routeMethods)[number][];
};

export const createRouteFile = ({
  path = "defaultPath",
  targetType = "instant-response",
  targetConfigurationStatus = "200",
  enabledMethods = [...routeMethods],
}: CreateRouteFileParams = {}) => {
  const methodsYaml = enabledMethods
    .map((method) => `${method}: { responses: { "200": { description: "" } } }`)
    .join("\n");

  return `direktiv_api: "endpoint/v2"
path: ${path}
allow_anonymous: true
plugins:
  inbound: []
  outbound: []
  auth: []
  target:
    type: ${targetType}
    configuration:
      status_code: ${targetConfigurationStatus}
${methodsYaml ? "\n" + methodsYaml : ""}`;
};

export const routeWithAWarning = `direktiv_api: "endpoint/v1"
timeout: 10000
allow_anonymous: true
plugins:
  inbound: []
  outbound: []
  auth: []
connect: { responses: { "200": { description: "" } } }
delete: { responses: { "200": { description: "" } } }`;

export const routeWithAnError = `direktiv_api: "endpoint/v1"
allow_anonymous: true
path: "test"
timeout: 10000
plugins:
  target:
    type: "this-plugin-does-not-exist"
    configuration:
      status_code: 200
      status_message: "Test"
  inbound: []
  outbound: []
  auth: []
`;

type FindRouteWithApiRequestParams = {
  namespace: string;
  match: (route: NewRouteSchemaType) => boolean;
};

type ErrorType = { response: { status?: number } };

export const findRouteWithApiRequest = async ({
  namespace,
  match,
}: FindRouteWithApiRequestParams) => {
  try {
    const { data: routes } = await getRoutes({
      urlParams: {
        baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
        namespace,
      },
      headers,
    });

    const normalizedRoutes = routes.map((route) => {
      const config = route.spec["x-direktiv-config"];
      return {
        ...route,
        spec: {
          ...route.spec,
          "x-direktiv-config": {
            ...config,
            plugins: {
              ...config.plugins,
              inbound: config.plugins.inbound ?? [],
              outbound: config.plugins.outbound ?? [],
              auth: config.plugins.auth ?? [],
              target: config.plugins.target,
            },
          },
        },
      };
    });
    return normalizedRoutes.find(match);
  } catch (error) {
    const typedError = error as ErrorType;
    if (typedError.response.status === 404) {
      return false;
    }
    throw new Error(
      `Unexpected error ${typedError?.response?.status} during lookup of service ${match} in namespace ${namespace}`
    );
  }
};
