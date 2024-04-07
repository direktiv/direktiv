const groupsYaml = (groups: string[]) =>
  groups.map((group) => `\n  - ${group}`).join("");

const tagsYaml = (tags: string[]) => tags.map((tag) => `\n  - ${tag}`).join("");

type Consumer = {
  username: string;
  password: string;
  apiKey: string;
  tags: string[];
  groups: string[];
};

export const createConsumerYaml = ({
  username,
  password,
  apiKey,
  groups,
  tags,
}: Consumer) => `direktiv_api: consumer/v1
username: ${username}
password: ${password}
api_key: ${apiKey}
tags:${tagsYaml(tags)}
groups:${groupsYaml(groups)}
`;
