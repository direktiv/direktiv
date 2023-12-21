import { EndpointFormSchema, EndpointFormSchemaType } from "./schema";

import { ZodError } from "zod";
import { stringify as jsonToPrettyYamlStringify } from "json-to-pretty-yaml";
import yamljs from "js-yaml";

type ReturnType =
  | [EndpointFormSchemaType, undefined]
  | [undefined, ZodError<EndpointFormSchemaType>];

export const serializeEndpointFile = (yaml: string): ReturnType => {
  let json;
  try {
    json = yamljs.load(yaml);
  } catch (e) {
    json = null;
  }

  const jsonParsed = EndpointFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultEndpointFileJson: EndpointFormSchemaType = {
  direktiv_api: "endpoint/v1",
};

export const defaultEndpointFileYaml = jsonToPrettyYamlStringify(
  defaultEndpointFileJson
);

/**
 * a wrapper around the stringify method of json-to-pretty-yaml
 * but it will serialize an empty object to an empty string instead
 * of "{}"
 */
export const jsonToYaml = (t: Record<string, unknown>) =>
  Object.keys(t).length === 0 ? "" : jsonToPrettyYamlStringify(t);
