import { z } from "zod";

export const StatusSchema = z.enum(["True", "False", "Unknown"]);

export type StatusSchemaType = z.infer<typeof StatusSchema>;

/**
 * example
  {
    "type": "UpAndReady",
    "status": "True",
    "message": "Up 4 days"
  }
 */
const ConditionSchema = z.object({
  type: z.string(),
  status: StatusSchema,
  message: z.string().optional(),
});

/**
  * example
  {
    "id": "obj949dad869e2ef05dbf77obj",
    "type": "namespace-service",
    "namespace": "test",
    "name": "s1",
    "filePath": "/service-test.yaml",
    "image": "redis",
    "cmd": "redis-server",
    "size": "",
    "scale": 0,
    "error": null,
    "conditions": [...]
  },
 */
const ServiceSchema = z.object({
  id: z.string(),
  type: z.enum(["namespace-service", "workflow-service"]),
  namespace: z.string(),
  name: z.string().nullable(),
  filePath: z.string(),
  image: z.string(),
  cmd: z.string(),
  size: z.string(),
  scale: z.number(),
  error: z.string().nullable(),
  conditions: z.array(ConditionSchema).nullable(),
});

/**
 * example
  {
    "data": [{...}, {...}, {...}]
  } 
 */
export const ServicesListSchema = z.object({
  data: z.array(ServiceSchema),
});

export const ServiceRebuildSchema = z.null();

export type ServiceSchemaType = z.infer<typeof ServiceSchema>;
export type ServicesListSchemaType = z.infer<typeof ServicesListSchema>;
