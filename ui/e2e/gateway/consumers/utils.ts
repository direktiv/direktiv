import { ConsumerSchemaType } from "~/api/gateway/schema";
import { getConsumers } from "~/api/gateway/query/getConsumers";
import { headers } from "e2e/utils/testutils";

type CreateRedisConsumerFileParams = {
  username?: string;
  password?: string;
};

export const createRedisConsumerFile = ({
  username = "userA",
  password = "password",
}: CreateRedisConsumerFileParams = {}) => `direktiv_api: consumer/v1
username: ${username}
password: ${password}
api_key: "123456789"
groups: 
- "group1"
- "group2"
tags:
  - "tag1"
`;

type FindConsumerWithApiRequestParams = {
  namespace: string;
  match: (consumer: ConsumerSchemaType) => boolean;
};

type ErrorType = { response: { status?: number } };

export const findConsumerWithApiRequest = async ({
  namespace,
  match,
}: FindConsumerWithApiRequestParams) => {
  try {
    const { data: consumers } = await getConsumers({
      urlParams: {
        baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
        namespace,
      },
      headers,
    });
    return consumers.find(match);
  } catch (error) {
    const typedError = error as ErrorType;
    if (typedError.response.status === 404) {
      // fail silently to allow for using poll() in tests
      return false;
    }
    throw new Error(
      `Unexpected error ${typedError?.response?.status} during lookup of consumer ${match} in namespace ${namespace}`
    );
  }
};
