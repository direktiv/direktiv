import { EndpointLoadSchema, EndpointLoadSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { ZodError } from "zod";

type SerializeReturnType =
  | [ReturnType<typeof EndpointLoadSchema.parse>, undefined]
  | [undefined, ZodError];

export const serializeEndpointFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

  const parsed = EndpointLoadSchema.safeParse(json);
  if (parsed.success) {
    return [parsed.data, undefined];
  }
  return [undefined, parsed.error];
};

const defaultEndpointFileJson: EndpointLoadSchemaType = {
  "x-direktiv-api": "endpoint/v2",
  "x-direktiv-config": {},
};

export const defaultEndpointFileYaml = jsonToYaml(defaultEndpointFileJson);
