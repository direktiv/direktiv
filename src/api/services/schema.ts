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
  conditions: z.array(ServiceRevisionConditionSchema).optional(),
  revision: z.string(),
  minScale: z.number().optional(),
  size: z.number().optional(),
});

// streaming violates the schema at two fields, so we create a new
// schema for streaming that will ignore this fields, when updating
// the cache, we will not update these fields (will not change anyways)
const ServiceRevisionSchemaStreaming = ServiceRevisionSchema.omit({
  created: true, // created is a string when received via streaming
  revision: true, // not present when streamed 🫠
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

export const ServiceRevisionStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  revision: ServiceRevisionSchemaStreaming,
});

/**
 * example
  {
    "name": "namespace-14529307612894023951-00004",
    "image": "gcr.io/direktiv/functions/hello-world:1.0",
    "cmd": "",
    "size": 1,
    "minScale": 1,
    "generation": "0",
    "created": "1692342131",
    "status": "True",
    "conditions": [
      {
        "name": "Active",
        "status": "True",
        "reason": "",
        "message": ""
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
    "desiredReplicas": "1",
    "actualReplicas": "1",
    "rev": "00004"
  }
 */
export const ServiceRevisionDetailSchema = z.object({
  name: z.string(),
  image: z.string(),
  cmd: z.string(),
  size: z.number(),
  minScale: z.number(),
  generation: z.string(),
  created: z.string(),
  status: StatusSchema,
  conditions: z.array(ServiceRevisionConditionSchema),
  desiredReplicas: z.string(),
  actualReplicas: z.string(),
  rev: z.string(),
});

export const ServiceRevisionDetailStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  revision: ServiceRevisionDetailSchema,
});

/**
 * example
  {
    "name": "namespace-14529307612894023951-00004-deployment-76d465f47cqvfk7",
    "status": "Running",
    "serviceName": "namespace-14529307612894023951",
    "serviceRevision": "namespace-14529307612894023951-00004"
  }
 */

export const PodsSchema = z.object({
  name: z.string(),
  status: z.string(), // TODO: make enum
  serviceName: z.string(),
  serviceRevision: z.string(),
});

/**
 * example
  {
    "pods": []
  }
 */
export const PodsListSchema = z.object({
  pods: z.array(PodsSchema),
});

export const PodsStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  pod: PodsSchema,
});

/**
 * example
  {
    "data": "2023/08/18 07:02:13 Serving hello world at http://[::]:8080\n"
  }
 */
export const PodLogsSchema = z.object({
  data: z.string(),
});

export const ServiceDeletedSchema = z.null();

export const ServiceCreatedSchema = z.null();

export const ServiceRevisionCreatedSchema = z.null();

export const ServiceRevisionDeletedSchema = z.null();

export type ServiceSchemaType = z.infer<typeof ServiceSchema>;
export type StatusSchemaType = z.infer<typeof StatusSchema>;
export type ServicesListSchemaType = z.infer<typeof ServicesListSchema>;
export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
export type ServiceStreamingSchemaType = z.infer<typeof ServiceStreamingSchema>;
export type ServiceRevisionStreamingSchemaType = z.infer<
  typeof ServiceRevisionStreamingSchema
>;
export type ServiceRevisionSchemaType = z.infer<typeof ServiceRevisionSchema>;
export type ServicesRevisionListSchemaType = z.infer<
  typeof ServicesRevisionListSchema
>;
export type ServiceRevisionFormSchemaType = z.infer<
  typeof ServiceRevisionFormSchema
>;

export type ServiceRevisionDetailSchemaType = z.infer<
  typeof ServiceRevisionDetailSchema
>;
export type ServiceRevisionDetailStreamingSchemaType = z.infer<
  typeof ServiceRevisionDetailStreamingSchema
>;

export type PodsListSchemaType = z.infer<typeof PodsListSchema>;
export type PodsSchemaType = z.infer<typeof PodsSchema>;
export type PodsStreamingSchemaType = z.infer<typeof PodsStreamingSchema>;

export type PodLogsSchemaType = z.infer<typeof PodLogsSchema>;
