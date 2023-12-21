import { stringify as jsonToPrettyYamlStringify } from "json-to-pretty-yaml";
/**
 * a wrapper around the stringify method of json-to-pretty-yaml
 * but it will serialize an empty object to an empty string instead
 * of "{}"
 */
export const jsonToYaml = (t: Record<string, unknown>) =>
  Object.keys(t).length === 0 ? "" : jsonToPrettyYamlStringify(t);
