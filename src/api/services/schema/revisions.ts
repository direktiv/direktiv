import { SizeSchema, StatusSchema } from ".";

import { z } from "zod";

export const revisionConditionNames = [
  "Active",
  "ContainerHealthy",
  "Ready",
  "ResourcesAvailable",
] as const;

const RevisionConditionSchema = z
  .object({
    name: z.enum(revisionConditionNames),
    status: StatusSchema,
    reason: z.string(),
    message: z.string(),
  })
  /**
   * When the name is "Active" and the reason is "NoTraffic",
   * the status should be changed to "True". The backend
   * reports this state as an "error", but it should
   * not visually be an error, so we change it here
   */
  .transform((data) => ({
    ...data,
    status:
      data.status === "False" &&
      data.name === "Active" &&
      data.reason === "NoTraffic"
        ? "True"
        : data.status,
  }));

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
const RevisionSchema = z.object({
  name: z.string(),
  image: z.string(),
  created: z.number().or(z.string()),
  status: StatusSchema,
  conditions: z.array(RevisionConditionSchema).optional(),
  revision: z.string().optional(),
  rev: z.string().optional(),
  minScale: z.number().optional(),
  size: SizeSchema.optional(),
});

export const RevisionFormSchema = z.object({
  cmd: z.string(),
  image: z.string().nonempty(),
  size: SizeSchema,
  // scale also has a max value, but it is dynamic depending on the namespace
  minscale: z.number().int().gte(0),
});

export const RevisionCreatedSchema = z.null();

export const RevisionDeletedSchema = z.null();

// streaming violates the schema at two fields, so we create a new
// schema for streaming that will ignore these fields, when updating
// the cache, we will not update these fields (will not change anyways)
const RevisionSchemaWhenStreamed = RevisionSchema.omit({
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
export const RevisionsListSchema = z.object({
  name: z.string().optional(),
  config: z
    .object({
      maxscale: z.number(),
    })
    .optional(),
  revisions: z.array(RevisionSchema).optional(),
});

export const RevisionStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  revision: RevisionSchemaWhenStreamed,
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
export const RevisionDetailSchema = z.object({
  name: z.string(),
  image: z.string(),
  cmd: z.string(),
  size: z.number(),
  minScale: z.number(),
  generation: z.string(),
  created: z.string(),
  status: StatusSchema,
  conditions: z.array(RevisionConditionSchema),
  desiredReplicas: z.string(),
  actualReplicas: z.string(),
  rev: z.string(),
});

export const RevisionDetailStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  revision: RevisionDetailSchema,
});

export type RevisionSchemaType = z.infer<typeof RevisionSchema>;
export type RevisionsListSchemaType = z.infer<typeof RevisionsListSchema>;
export type RevisionFormSchemaType = z.infer<typeof RevisionFormSchema>;
export type RevisionStreamingSchemaType = z.infer<
  typeof RevisionStreamingSchema
>;

export type RevisionDetailSchemaType = z.infer<typeof RevisionDetailSchema>;
export type RevisionDetailStreamingSchemaType = z.infer<
  typeof RevisionDetailStreamingSchema
>;
