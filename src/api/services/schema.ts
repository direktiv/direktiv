import { z } from "zod";

export const StatusSchema = z.enum(["True", "False", "Unknown"]);

export const serviceConditionNames = [
  "ConfigurationsReady",
  "Ready",
  "RoutesReady",
] as const;

const ServiceConditionSchema = z.object({
  name: z.enum(serviceConditionNames),
  status: StatusSchema,
  reason: z.string(),
  message: z.string(),
});

export const serviceRevisionConditionNames = [
  "Active",
  "ContainerHealthy",
  "Ready",
  "ResourcesAvailable",
] as const;

const ServiceRevisionConditionSchema = z.object({
  name: z.enum(serviceRevisionConditionNames),
  status: StatusSchema,
  reason: z.string(),
  message: z.string(),
});

/**
 * example
  {
    "info": {
      "name": "service",
      "namespace": "c75454f2-3790-4f36-a1a2-22ca8a4f8020",
      "workflow": "",
      "image": "direktiv/request",
      "cmd": "",
      "size": 0,
      "minScale": 0,
      "namespaceName": "stefan",
      "path": "",
      "revision": "",
      "envs": {}
    },
    "status": "True",
    "conditions": [
      {
        "name": "ConfigurationsReady",
        "status": "True",
        "reason": "",
        "message": ""
      },
      {
        "name": "Ready",
        "status": "True",
        "reason": "",
        "message": ""
      },
      {
        "name": "RoutesReady",
        "status": "True",
        "reason": "",
        "message": ""
      }
    ],
    "serviceName": "namespace-14895841056527822151"
  }
 */
const ServiceSchema = z.object({
  info: z.object({
    name: z.string(),
    namespace: z.string(),
    workflow: z.string(),
    image: z.string(), // direktiv/request"
    cmd: z.string(),
    size: z.number(),
    minScale: z.number(),
    namespaceName: z.string(),
    path: z.string(),
    revision: z.string(),
    envs: z.object({}),
  }),
  status: StatusSchema,
  conditions: z.array(ServiceConditionSchema),
  serviceName: z.string(),
});

/**
 * example
  {
    "config": {
      "maxscale": 3
    },
    "functions": []
  }
 */
export const ServicesListSchema = z.object({
  config: z.object({
    maxscale: z.number(),
  }),
  functions: z.array(ServiceSchema),
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
  minscale: z.number().int().gte(0).lte(3),
  // scale also has a max value, but it is dynamic depending on the namespace
  size: z.number().int().gte(0),
});

export const ServiceRevisionFormSchema = z.object({
  cmd: z.string(),
  image: z.string().nonempty(),
  minscale: z.number().int().gte(0).lte(3),
  // scale also has a max value, but it is dynamic depending on the namespace
  size: z.number().int().gte(0),
});

/**
 * example
  {
    "name": "namespace-14937757830533003475-00001",
    "image": "direktiv/solve:v3",
    "created": 1691140028,
    "status": "True",
    "minScale" : 1,
    "size" : 1,
    "conditions": [
      {
        "name": "Active",
        "status": "False",
        "reason": "NoTraffic",
        "message": "The target is not receiving traffic."
      },
      {
        "name": "ContainerHealthy",
        "status": "True",
        "reason": "",
        "message": ""
      },
      {
        "name": "Ready",
        "status": "True",
        "reason": "",
        "message": ""
      },
      {
        "name": "ResourcesAvailable",
        "status": "True",
        "reason": "",
        "message": ""
      }
    ],
    "revision": "00001"
  }

 */
const ServiceRevisionSchema = z.object({
  name: z.string(),
  image: z.string(),
  created: z.number(),
  status: StatusSchema,
  conditions: z.array(ServiceRevisionConditionSchema),
  revision: z.string(),
  minScale: z.number().optional(),
  size: z.number().optional(),
});

/**
 * example
  {
    "name": "name123",
    "namespace": "sebxian",
    "config": {
      "maxscale": 3
    },
    "revisions": [],
    "scope": "namespace"
  }
 */
export const ServicesRevisionListSchema = z.object({
  name: z.string(),
  config: z.object({
    maxscale: z.number(),
  }),
  revisions: z.array(ServiceRevisionSchema).optional(),
});

export const ServiceDeletedSchema = z.null();

export const ServiceCreatedSchema = z.null();

export const ServiceRevisionCreatedSchema = z.null();

export type ServiceSchemaType = z.infer<typeof ServiceSchema>;
export type StatusSchemaType = z.infer<typeof StatusSchema>;
export type ServicesListSchemaType = z.infer<typeof ServicesListSchema>;
export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
export type ServiceStreamingSchemaType = z.infer<typeof ServiceStreamingSchema>;
export type ServiceRevisionFormSchemaType = z.infer<
  typeof ServiceRevisionFormSchema
>;
