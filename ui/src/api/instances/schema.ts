import { z } from "zod";

const InstanceStatusSchema = z.enum([
  "pending",
  "failed",
  "cancelled",
  "crashed",
  "complete",
]);

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
const parentInstance = z.object({
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
    "createdAt": "0001-01-01T00:00:00Z",
    "endedAt": null,
    "status": "pending",
    "path": "/test.yaml",
    "errorCode": "",
    "invoker": "api",
    "definition": "ZGlyZWt0aXZfYXBpOiB3b3JrZmxvdy92MQpkZXNjcmlwdGlvbjogQSBzaW1wbGUgJ25vLW9wJyBzdGF0ZSB0aGF0IHJldHVybnMgJ0hlbGxvIHdvcmxkIScKc3RhdGVzOgotIGlkOiBoZWxsb3dvcmxkCiAgdHlwZTogbm9vcAogIHRyYW5zZm9ybToKICAgIHJlc3VsdDogSGVsbG8gd29ybGQhCg==",
    "flow": [],
    "traceId": "00000000000000000000000000000000",
    "lineage": [...]
  }
 */

const InstanceSchema = z.object({
  id: z.string(),
  createdAt: z.string(),
  endedAt: z.string().nullable(),
  status: InstanceStatusSchema,
  path: z.string(),
  errorCode: z.string().nullable(),
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
  lineage: z.array(parentInstance),
});

/**
 * example
 * 
  {
    "data": {...}
  }
 */
export const InstanceCreateSchema = z.object({
  data: InstanceSchema,
});

export const InstanceCancelPayload = z.object({
  status: z.literal("cancelled"),
});

export type InstanceCancelPayloadType = z.infer<typeof InstanceCancelPayload>;

export const InstanceCancelSchema = z.null();

export const InstancesListSchema = z.object({
  data: z.array(InstanceSchema),
});
