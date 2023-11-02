import { z } from "zod";

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
  configuration: z.record(z.unknown()),
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
  method: z.enum(["GET", "POST", "PUT", "DELETE", "PATCH", ""]),
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

export type GatewaySchemeType = z.infer<typeof EndpointSchema>;
