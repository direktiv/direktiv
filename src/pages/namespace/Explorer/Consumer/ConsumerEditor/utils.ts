import { stringify } from "json-to-pretty-yaml";

export const consumerHeader = {
  direktiv_api: "consumer/v1",
};

export const defaultConsumerYaml = stringify(consumerHeader);
