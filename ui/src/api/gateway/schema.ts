import { z } from "zod";

// x-direktiv-api: endpoint/v2

// x-direktiv-config:
//   path: "/testme"
//   allow_anonymous: true
//   plugins:
//     target:
//       type: instant-response
//       configuration:
//         status_code: 200
//         status_message: hello
// get:
//   description: Optional extended description in CommonMark or HTML.
//   responses:
//     "200":
//       description: returns something

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

export const OperationSchema = z.object({
  description: z.string().optional(),
  responses: z.record(z.any()).optional(),
});

// I am helping TypeScript preserve the literal key types for
// when being spread into the NewRouteSchema
export type RouteMethod = (typeof routeMethods)[number];

export const methodSchemas = routeMethods.reduce<
  Record<RouteMethod, z.ZodTypeAny>
>(
  (acc, method) => {
    acc[method] = OperationSchema.optional();
    return acc;
  },
  {} as Record<RouteMethod, z.ZodTypeAny>
);

export const NewRouteSchema = z.object({
  spec: z.object({
    "x-direktiv-api": z.literal("endpoint/v2"),
    "x-direktiv-config": z.object({
      allow_anonymous: z.boolean(),
      path: z.string(),
      plugins: z
        .object({
          inbound: filterInvalidEntries(PluginSchema).optional(),
          outbound: filterInvalidEntries(PluginSchema).optional(),
          auth: filterInvalidEntries(PluginSchema).optional(),
          target: PluginSchema.optional(),
        })
        .optional(),
    }),
    ...methodSchemas,
  }),
  file_path: z.string(),
  errors: z.array(z.string()),
  server_path: z.string(),
  warnings: z.array(z.string()),
});

export type RouteSchemaType = z.infer<typeof NewRouteSchema>;

export type MethodsKeys = keyof RouteSchemaType["spec"];

export type MethodsObject = Partial<Pick<RouteSchemaType["spec"], MethodsKeys>>;
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
