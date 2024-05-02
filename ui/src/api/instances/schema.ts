import { z } from "zod";

export const possibleInstanceStatuses = [
  "pending",
  "failed",
  "cancelled",
  "crashed",
  "complete",
] as const;

const InstanceStatusSchema = z.enum(possibleInstanceStatuses);

export const possibleTriggerValues = [
  "api",
  "cloudevent",
  "instance",
  "cron",
] as const;

/**
 * example
 * 
  {
    "branch": 0,
    "id": "77f369d9-085a-4a7b-9695-c2ec7a469071",
    "state": "getter",
    "step": 9
  }
 */
const ParentInstanceSchema = z.object({
  branch: z.number(),
  id: z.string(),
  state: z.string(),
  step: z.number(),
});

/**
 * example
 * 
  {
    "id": "79310904-929f-4f83-bed6-0bf8c1e49dc1",
    "namespace": "test",
    "createdAt": "0001-01-01T00:00:00Z",
    "endedAt": null,
    "status": "pending",
    "path": "/test.yaml",
    "errorCode": "",
    "invoker": "api",
    "definition": "ZGlyZWt0aXZfYXBpOiB3b3JrZmxvdy92MQpkZXNjcmlwdGlvbjogQSBzaW1wbGUgJ25vLW9wJyBzdGF0ZSB0aGF0IHJldHVybnMgJ0hlbGxvIHdvcmxkIScKc3RhdGVzOgotIGlkOiBoZWxsb3dvcmxkCiAgdHlwZTogbm9vcAogIHRyYW5zZm9ybToKICAgIHJlc3VsdDogSGVsbG8gd29ybGQhCg==",
    "errorMessage":"c3ViamVjdCBmYWlsZWQgaXRzIEpTT05TY2hlbWEgdmFsaWRhdGlvbjogPG5pbD4=",
    "flow": [ "prep", "loop", "getter" ],
    "traceId": "00000000000000000000000000000000",
    "lineage": [...]
  }
 */
const InstanceSchema = z.object({
  id: z.string(),
  namespace: z.string(),
  createdAt: z.string(),
  endedAt: z.string().nullable(),
  status: InstanceStatusSchema,
  path: z.string(),
  errorCode: z.string().nullable(),
  errorMessage: z.string().nullable(),
  /**
   * either "api", "cron", "cloudevent" or "complete"
   * if it's created as a subflow from another instance
   * it's something like instance:%v, where %v is the
   * instance ID of its parent
   */
  invoker: z.string(),
  definition: z.string(),
  flow: z.array(z.string()),
  traceId: z.string().nullable(),
  lineage: z.array(ParentInstanceSchema),
});

export type InstanceSchemaType = z.infer<typeof InstanceSchema>;

/**
 * example
 * 
  {
    "data": {...}
  }
 */
export const InstanceCreatedResponseSchema = z.object({
  data: InstanceSchema,
});

export const InstanceCancelPayload = z.object({
  status: z.literal("cancelled"),
});

export type InstanceCancelPayloadType = z.infer<typeof InstanceCancelPayload>;

export const InstanceCanceledResponseSchema = z.null();

/**
 * example
 * 
  { 
    "meta": {
      "total": 278,
    },
    "data": {...}
  }
 */
export const InstancesListResponseSchema = z.object({
  meta: z.object({
    total: z.number(),
  }),
  data: z.array(InstanceSchema),
});

/**
 * example
  {
    ...
    "inputLength" : 8,
    "metadataLength" : 0,
    "outputLength" : 7,
  } 
 */
export const InstanceDetailsSchema = InstanceSchema.extend({
  inputLength: z.number(),
  outputLength: z.number(),
  metadataLength: z.number(),
});

export const InstanceDetailsResponseSchema = z.object({
  data: InstanceDetailsSchema,
});

export type InstanceDetailsResponseSchemaType = z.infer<
  typeof InstanceDetailsResponseSchema
>;

/**
 * example
 * 
  {
    ... 
    "inputLength": 8,
    "input": "ewogICAgCn0="
  }
 */
export const InstancesInputResponseSchema = z.object({
  data: InstanceSchema.extend({
    inputLength: z.number(),
    input: z.string(),
  }),
});

/**
 * example
 * 
  {
    ... 
    "outputLength": 7,
    "output": "eyJ4IjowfQ=="
  }
 */
export const InstancesOutputResponseSchema = z.object({
  data: InstanceSchema.extend({
    outputLength: z.number(),
    output: z.string().optional(),
  }),
});
