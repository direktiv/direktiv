import { ConsumerFormSchema, ConsumerFormSchemaType } from "./schema";

import { ZodError } from "zod";
import { stringify as jsonToPrettyYamlStringify } from "json-to-pretty-yaml";
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

export const defaultConsumerFileYaml = jsonToPrettyYamlStringify(
  defaultConsumerFileJson
);

/**
 * a wrapper around the stringify method of json-to-pretty-yaml
 * but it will serialize an empty object to an empty string instead
 * of "{}"
 *
 * TODO: merge this function with the one in EndpointEditor/utils.ts
 */
export const jsonToYaml = (t: Record<string, unknown>) =>
  Object.keys(t).length === 0 ? "" : jsonToPrettyYamlStringify(t);
