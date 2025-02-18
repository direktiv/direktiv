import { z } from "zod";

export const routeMethods = [
  "connect",
  "delete",
  "get",
  "head",
  "options",
  "patch",
  "post",
  "put",
  "trace",
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
  configuration: z.record(z.unknown()).optional(),
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

export const NewRouteSchema = z.object({
  spec: z.object({
    "x-direktiv-api": z.literal("endpoint/v2"),
    "x-direktiv-config": z.object({
      allow_anonymous: z.boolean(),
      methods: filterInvalidEntries(MethodsSchema).nullable(),
      path: z.string(),
      plugins: z.object({
        inbound: filterInvalidEntries(PluginSchema).default([]),
        outbound: filterInvalidEntries(PluginSchema).default([]),
        auth: filterInvalidEntries(PluginSchema).default([]),
        target: PluginSchema,
      }),
    }),
  }),
  file_path: z.string(),
  errors: z.array(z.string()),
  server_path: z.string(),
  warnings: z.array(z.string()),
});

export type NewRouteSchemaType = z.infer<typeof NewRouteSchema>;

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const RoutesListSchema = z.object({
  data: z.array(NewRouteSchema),
});

export type RoutesListSchemaType = z.infer<typeof RoutesListSchema>;

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
  tags: z.array(z.string()).nullable(),
  groups: z.array(z.string()).nullable(),
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

/**
 * example INFO
{
  "data": {
    "spec": {
      "openapi": "3.0.0",
      "info": {
        "title": "cg",
        "version": "1.0"
      },
      "paths": {}
    },
    "file_path": "virtual",
    "errors": []
  }
}
 */
export const OpenapiSpecificationSchema = z.object({
  data: z.object({
    spec: z
      .object({
        openapi: z.string(),
        info: z
          .object({
            title: z.string(),
            version: z.string(),
            description: z.string().optional(),
          })
          .passthrough(),
        paths: z.record(z.any()),
      })
      .passthrough(),
    file_path: z.string(),
    errors: z.array(z.string()),
  }),
});

export type OpenapiSpecificationSchemaType = z.infer<
  typeof OpenapiSpecificationSchema
>;
