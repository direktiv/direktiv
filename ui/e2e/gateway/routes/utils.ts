import { NewRouteSchemaType } from "~/api/gateway/schema";
import { getRoutes } from "~/api/gateway/query/getRoutes";
import { headers } from "e2e/utils/testutils";

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
  - "DELETE"
  - "OPTIONS"
  - "PUT"
  - "POST"
  - "HEAD"
  - "CONNECT"
  - "PATCH"
  - "TRACE"
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
methods: []
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
    return routes.find(match);
  } catch (error) {
    // In case tests act up due to unexpected errors, we can use the following
    // line (and comment the below) to fail silently on all errors.
    // return false;
    const typedError = error as ErrorType;
    if (typedError.response.status === 404) {
      // fail silently to allow for using poll() in tests
      return false;
    }
    throw new Error(
      `Unexpected error ${typedError?.response?.status} during lookup of service ${match} in namespace ${namespace}`
    );
  }
};
