import { ConsumerFormSchema, ConsumerFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { ZodError } from "zod";

type SerializeReturnType =
  | [ConsumerFormSchemaType, undefined]
  | [undefined, ZodError<ConsumerFormSchemaType>];

export const serializeConsumerFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

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
