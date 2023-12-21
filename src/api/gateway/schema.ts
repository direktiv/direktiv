import { z } from "zod";

export const routeMethods = [
  "CONNECT",
  "DELETE",
  "GET",
  "HEAD",
  "OPTIONS",
  "PATCH",
  "POST",
  "PUT",
  "TRACE",
] as const;

export const MethodsSchema = z.enum(routeMethods);

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

const filterInvalidEntries = (schema: z.ZodTypeAny) =>
  z
    .array(z.any())
    .transform((entryArr) =>
      entryArr.filter((entry) => schema.safeParse(entry).success)
    );

const RouteSchema = z.object({
  methods: filterInvalidEntries(MethodsSchema).nullable(),
  file_path: z.string(),
  path: z.string().optional(),
  server_path: z.string().optional(),
  allow_anonymous: z.boolean(),
  errors: z.array(z.string()),
  warnings: z.array(z.string()),
  plugins: z.object({
    // if a user might use an unsupported plugin, they will be parsed out instead of throwing an error
    outbound: filterInvalidEntries(PluginSchema).optional(),
    inbound: filterInvalidEntries(PluginSchema).optional(),
    auth: filterInvalidEntries(PluginSchema).optional(),
    target: PluginSchema.optional(),
  }),
});

export type RouteSchemeType = z.infer<typeof RouteSchema>;

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

export type ConsumerSchemaType = z.infer<typeof ConsumerSchema>;

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const ConsumersListSchema = z.object({
  data: z.array(ConsumerSchema),
});
