import { ConsumerFormSchema, ConsumerFormSchemaType } from "./schema";

import { ZodError } from "zod";
import { jsonToYaml } from "../../utils";
import yamljs from "js-yaml";

type SerializeReturnType =
  | [ConsumerFormSchemaType, undefined]
  | [undefined, ZodError<ConsumerFormSchemaType>];

export const serializeConsumerFile = (yaml: string): SerializeReturnType => {
  let json;
  try {
    json = yamljs.load(yaml);
  } catch (e) {
    json = null;
  }

  const jsonParsed = ConsumerFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultConsumerFileJson: ConsumerFormSchemaType = {
  direktiv_api: "consumer/v1",
};

export const defaultConsumerFileYaml = jsonToYaml(defaultConsumerFileJson);
