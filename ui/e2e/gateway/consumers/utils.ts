import { ConsumerSchemaType } from "~/api/gateway/schema";
import { getConsumers } from "~/api/gateway/query/getConsumers";

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

export const findConsumerWithApiRequest = async ({
  namespace,
  match,
}: FindConsumerWithApiRequestParams) => {
  const { data: consumers } = await getConsumers({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
    },
  });
  return consumers.find(match);
};
