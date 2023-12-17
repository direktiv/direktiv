import { EndpointFormSchema, EndpointFormSchemaType } from "./schema";

import { ZodError } from "zod";
import { stringify } from "json-to-pretty-yaml";
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

export const defaultEndpointFileYaml = stringify(defaultEndpointFileJson);
