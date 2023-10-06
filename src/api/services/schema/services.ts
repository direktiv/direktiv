import { SizeSchema, StatusSchema } from ".";

import { z } from "zod";

export const serviceConditionNames = [
  "ConfigurationsReady",
  "Ready",
  "RoutesReady",
] as const;

/**
 * example
  {
    "type": "ConfigurationsReady",
    "status": "True",
    "lastTransitionTime": "2023-10-06T07:19:38Z"
  }
 */
const ConditionSchema = z.object({
  type: z.enum(serviceConditionNames),
  status: StatusSchema,
  lastTransitionTime: z.string(),
});

/**
  * example
  {
    "id": "objb50e89ac0959d41d7b08obj",
    "config": {
      "namespace": "mynamespace",
      "name": null,
      "servicePath": "/my-redis-service.yaml",
      "workflowPath": null,
      "image": "redis",
      "cmd": "",
      "size": "",
      "scale": 0,
      "error": null
    },
    "conditions": {...}
  }
 */
const ServiceSchema = z.object({
  id: z.string(),
  config: z.object({
    namespace: z.string(),
    name: z.string().nullable(),
    servicePath: z.string().nullable(),
    workflowPath: z.string().nullable(),
    image: z.string(),
    cmd: z.string(),
    size: z.string(),
    scale: z.number(),
    error: z.string().nullable(),
  }),
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
