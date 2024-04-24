import { ServiceSchemaType } from "~/api/services/schema/services";
import { getServices } from "~/api/services/query/services";

type CreateServiceFileParams = {
  scale?: number;
  size?: "large" | "medium" | "small";
};

export const createHttpServiceFile = ({
  scale = 1,
  size = "small",
}: CreateServiceFileParams = {}) => `direktiv_api: service/v1
image: gcr.io/direktiv/functions/http-request:1.0
scale: ${scale}
size: ${size}
cmd: 
envs:
  - name: MY_ENV_VAR
    value: env-var-value
`;

export const createRequestServiceFile = ({
  scale = 1,
  size = "small",
}: CreateServiceFileParams = {}) => `direktiv_api: service/v1
image: direktiv/request:v4
scale: ${scale}
size: ${size}
cmd: /request
envs:
  - name: MY_ENV_VAR
    value: env-var-value
`;

export const serviceWithAnError = `direktiv_api: service/v1
image: this-image-does-not-exist
scale: 1
size: small
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
    });
    return services.find(match);
  } catch (error) {
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
