import { z } from "zod";

export const endpointMethods = [
  "GET",
  "POST",
  "PUT",
  "DELETE",
  "PATCH",
] as const;

/**
 * example
  {
    "type": "proxy",
    "configuration": {
      "key1": ["value1", "value2", "value3"],
      "key2": "value3"
    }
  }
 */
const PluginSchema = z.object({
  type: z.string(),
  configuration: z.record(z.unknown()).nullable(),
});

/**
 * example
  {
    "method": "POST",
    "file_path": "/path-to-service.yaml",
    "workflow": "action.yaml",
    "namespace": "ns",
    "error": "Error some error message",
    "status": "failed"
    "plugins": [{...}, {...}, {...}],
  }
 */
const EndpointSchema = z.object({
  method: z.enum([...endpointMethods, ""]),
  file_path: z.string(),
  error: z.string(),
  plugins: z.array(PluginSchema),
});

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const EndpointListSchema = z.object({
  data: z.array(EndpointSchema),
});

/**
 * example
  {
    "$defs": {
      "examplePluginConfig": {
        "additionalProperties": false,
        "properties": {
          "echo_value": {
            "type": "string"
          }
        },
        "required": [
          "echo_value"
        ],
        "type": "object"
      }
    },
    "$id": "https://github.com/direktiv/direktiv/pkg/refactor/gateway/example-plugin-config",
    "$ref": "#/$defs/examplePluginConfig",
    "$schema": "https://json-schema.org/draft/2020-12/schema"
  } 
 */
const PluginJSONSchema = z.object({
  $defs: z.record(z.record(z.unknown())),
  $id: z.string(),
  $ref: z.string(),
  $schema: z.string(),
});

/**
 * example
  {
    "data": {
      "example_plugin": {...}
    }
  }
 */
export const PluginsListSchema = z.object({
  data: z.record(PluginJSONSchema),
});

export type PluginsListSchemaType = z.infer<typeof PluginsListSchema>;
export type PluginJSONSchemaType = z.infer<typeof PluginJSONSchema>;

export type GatewaySchemeType = z.infer<typeof EndpointSchema>;
