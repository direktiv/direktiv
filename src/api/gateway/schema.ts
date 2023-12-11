import { z } from "zod";

export const endpointMethods = [
  "GET",
  "DELETE",
  "HEAD",
  "POST",
  "PUT",
  "TRACE",
  "PATH",
  "OPTIONS",
  "CONNECT",
  "*",
] as const;

const MethodsSchema = z.enum(endpointMethods);

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
  "file_path": "/test/roottree.yaml",
  "path": "/adon/",
  "server_path": "/ns/git-test/adon",
  "methods": [
    "GET"
  ],
  "allow_anonymous": true,
  "timeout": 0,
  "errors": [],
  "warnings": [],
  "plugins": {
    "target": {
      "type": "instant-response",
      "configuration": {
        "content_type": "application/xml",
        "status_code": 200,
        "status_message": "..."
      }
    }
  }
}
 */
const RouteSchema = z.object({
  methods: z.array(MethodsSchema),
  file_path: z.string(),
  path: z.string(),
  allow_anonymous: z.boolean(),
  errors: z.array(z.string()),
  warnings: z.array(z.string()),
  plugins: z.object({
    outbound: z.array(PluginSchema).optional(),
    inbound: z.array(PluginSchema).optional(),
    auth: z.array(PluginSchema).optional(),
    EventTarget: z.array(PluginSchema).optional(),
  }),
});

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const RoutesListSchema = z.object({
  data: z.array(RouteSchema),
});

/**
 * example
{
  "data": [
    {
      "username": "user",
      "password": "pwd",
      "api_key": "key",
      "tags": [
        "tag1"
      ],
      "groups": [
        "group1"
      ]
    }
  ]
}
 */
const ConsumerSchema = z.object({
  username: z.string(),
  password: z.string(),
  api_key: z.string(),
  tags: z.array(z.string()),
  groups: z.array(z.string()),
});

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const ConsumersListSchema = z.object({
  data: z.array(ConsumerSchema),
});

export type GatewaySchemeType = z.infer<typeof RouteSchema>;
