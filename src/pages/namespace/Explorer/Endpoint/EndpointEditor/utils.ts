import { EndpointFormSchema, EndpointFormSchemaType } from "./schema";

import { stringify } from "json-to-pretty-yaml";
import yamljs from "js-yaml";

export const serializeEndpointFile = (yaml: string) => {
  let json;
  try {
    json = yamljs.load(yaml);
  } catch (e) {
    json = null;
  }

  const jsonParsed = EndpointFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return jsonParsed.data;
  }
  return undefined;
};

const defaultEndpointFileJson: EndpointFormSchemaType = {
  direktiv_api: "endpoint/v1",
};

export const defaultEndpointFileYaml = stringify(defaultEndpointFileJson);
