import { ServiceSchemaType } from "~/api/services/schema/services";
import { getServices } from "~/api/services/query/services";
import { headers } from "e2e/utils/testutils";

type CreateServiceFileParams = {
  scale?: number;
  size?: "large" | "medium" | "small";
};

export const createRequestServiceFile = ({
  scale = 1,
  size = "small",
}: CreateServiceFileParams = {}) => `{
  "image": "direktiv/request:v4",
  "scale": ${scale},
  "size": "${size}",
  "cmd": "/request",
  "envs": [
    {
      "name": "MY_ENV_VAR",
      "value": "env-var-value"
    }
  ]
}
`;

export const serviceWithAnError = `{
  "image": "nope",
  "scale": 1,
  "size": "small"
}
`;

type FindServiceWithApiRequestParams = {
  namespace: string;
  match: (service: ServiceSchemaType) => boolean;
};

type ErrorType = { response: { status?: number } };

export const findServiceWithApiRequest = async ({
  namespace,
  match,
}: FindServiceWithApiRequestParams) => {
  try {
    const { data: services } = await getServices({
      urlParams: {
        baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
        namespace,
      },
      headers,
    });
    // if no match, return null so .poll() will retry.
    return services.find(match) ?? null;
  } catch (error) {
    const typedError = error as ErrorType;
    if (typedError.response.status === 404) {
      // return null so .poll() will retry.
      return null;
    }
    throw new Error(
      `Unexpected error ${typedError?.response?.status} during lookup of service ${match} in namespace ${namespace}`
    );
  }
};
