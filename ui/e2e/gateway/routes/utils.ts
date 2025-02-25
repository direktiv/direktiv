import {
  RouteMethod,
  RouteSchemaType,
  routeMethods,
} from "~/api/gateway/schema";

import { getRoutes } from "~/api/gateway/query/getRoutes";
import { headers } from "e2e/utils/testutils";

type CreateRouteFileParams = {
  path?: string;
  targetType?: string;
  targetConfigurationStatus?: string;
  enabledMethods?: RouteMethod[];
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

  return `x-direktiv-api: endpoint/v2
x-direktiv-config:
  allow_anonymous: true
  path: ${path}
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

export const routeWithAWarning = `x-direktiv-api: endpoint/v2
x-direktiv-config:
  allow_anonymous: true
  path: defaultPath
  plugins:
    inbound: []
    outbound: []
    auth: []
delete: { responses: { "200": { description: "" } } }
options: { responses: { "200": { description: "" } } }`;

export const routeWithAnError = `x-direktiv-api: endpoint/v2
x-direktiv-config:
  allow_anonymous: true
  path: test
  plugins:
    target:
      type: "this-plugin-does-not-exist"
      configuration:
        status_code: 200
        status_message: "Test"
    inbound: []
    outbound: []
    auth: []`;

type FindRouteWithApiRequestParams = {
  namespace: string;
  match: (route: RouteSchemaType) => boolean;
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

    return routes.find(match);
  } catch (error) {
    const typedError = error as ErrorType;
    if (typedError.response.status === 404) {
      return false;
    }
    throw new Error(
      `Unexpected error ${typedError?.response?.status} during lookup of route ${match} in namespace ${namespace}`
    );
  }
};
