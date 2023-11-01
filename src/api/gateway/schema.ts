import { z } from "zod";

/**
 * example
  {
    "name": "proxy",
    "version": "1.0.0",
    "runtimeConfig": {
      "key1": ["value1", "value2", "value3"],
      "key2": "value3"
    }
  }
 */
const PluginSchema = z.object({
  name: z.string(),
  version: z.string(),
  runtimeConfig: z.record(z.unknown()),
});

/**
 * example
  {
    "path": "/path-to-service.yaml",
    "method": "POST",
    "timeoutSeconds": 30,
    "plugins": [{...}, {...}, {...}],
    "error": "Error some error message",
    "status": "failed"
  }
 */
const GatewaySchema = z.object({
  path: z.string(),
  method: z.enum(["GET", "POST", "PUT", "DELETE", "PATCH"]),
  timeoutSeconds: z.number(),
  plugins: z.array(PluginSchema),
  error: z.string().nullable(),
  status: z.enum(["healthy", "failed"]),
});

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const GatewayListSchema = z.object({
  data: z.array(GatewaySchema),
});
