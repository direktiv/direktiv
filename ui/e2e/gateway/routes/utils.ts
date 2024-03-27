import { RouteSchemaType } from "~/api/gateway/schema";
import { getRoutes } from "~/api/gateway/query/getRoutes";

type CreateRouteFileParams = {
  path?: string;
  targetType?: string;
  targetConfigurationStatus?: string;
};

export const createRouteFile = ({
  path = "defaultPath",
  targetType = "instant-response",
  targetConfigurationStatus = "200",
}: CreateRouteFileParams = {}) =>
  `direktiv_api: "endpoint/v1"
path: ${path}
methods:
  - "GET"
allow_anonymous: true
plugins:
  inbound: []
  outbound: []
  auth: []
  target:
    type: ${targetType}
    configuration:
      status_code: ${targetConfigurationStatus}`;

export const routeWithAWarning = `direktiv_api: "endpoint/v1"
timeout: 10000
methods:
  - "CONNECT"
  - "DELETE"
allow_anonymous: true
plugins:
  inbound: []
  outbound: []
  auth: []`;

export const routeWithAnError = `direktiv_api: "endpoint/v1"
allow_anonymous: true
path: "test"
timeout: 10000
methods:
  - "CONNECT"
  - "DELETE"
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
  match: (route: RouteSchemaType) => boolean;
};

export const findRouteWithApiRequest = async ({
  namespace,
  match,
}: FindRouteWithApiRequestParams) => {
  const { data: routes } = await getRoutes({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
    },
  });
  return routes.find(match);
};
