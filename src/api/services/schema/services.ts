import { SizeSchema, StatusSchema } from ".";

import { z } from "zod";

export const serviceConditionNames = ["UpAndReady"] as const;

/**
 * example
  {
    "type": "UpAndReady",
    "status": "True",
    "message": "Up 4 days"
  }
 */
const ConditionSchema = z.object({
  type: z.enum(serviceConditionNames),
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

/**
 * example
  {
    "event": "ADDED",
    "function": {
      ...
    },
    "traffic": []
  } 
 */
export const ServiceStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  function: ServiceSchema,
});

export const serviceNameSchema = z
  .string()
  .nonempty()
  .regex(/^[a-z]([-a-z0-9]{0,62}[a-z0-9])?$/, {
    message:
      "Please use a name that only contains lowercase letters, and use - instead of whitespaces.",
  });

export const ServiceFormSchema = z.object({
  name: serviceNameSchema,
  cmd: z.string(),
  image: z.string().nonempty(),
  size: SizeSchema,
  // scale also has a max value, but it is dynamic depending on the namespace
  minscale: z.number().int().gte(0),
});

export const ServiceDeletedSchema = z.null();

export const ServiceCreatedSchema = z.null();

export type ServiceSchemaType = z.infer<typeof ServiceSchema>;
export type ServicesListSchemaType = z.infer<typeof ServicesListSchema>;
export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
export type ServiceStreamingSchemaType = z.infer<typeof ServiceStreamingSchema>;
