import { ServiceSchemaType } from "~/api/services/schema/services";
import { getServices } from "~/api/services/query/services";

type CreateRedisServiceFileParam = {
  scale?: number;
  size?: "large" | "medium" | "small";
};

export const createRedisServiceFile = ({
  scale = 1,
  size = "small",
}: CreateRedisServiceFileParam = {}) => `direktiv_api: service/v1
image: "redis"
scale: ${scale}
size: ${size}
cmd: "redis-server"
envs:
  - name: "MY_ENV_VAR"
    value: "env-var-value"
`;

export const serviceWithAnError = `direktiv_api: service/v1
image: "this-image-does-not-exist"
scale: 1
size: "small"
`;

type FindServiceViaApiParam = {
  namespace: string;
  find: (service: ServiceSchemaType) => boolean;
};

export const findServiceViaApi = async ({
  namespace,
  find,
}: FindServiceViaApiParam) => {
  const { data: services } = await getServices({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
    },
  });

  return services.some(find);
};
