import { stringify } from "json-to-pretty-yaml";

export const endpointHeader = {
  direktiv_api: "endpoint/v1",
};

export const defaultEndpointYaml = stringify(endpointHeader);
