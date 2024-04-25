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
    "lineage": []
  }
 */
const InstanceSchema = z.object({
  id: z.string(),
  createdAt: z.string(),
  endedAt: z.string().nullable(),
  status: InstanceStatusSchema,
  path: z.string(),
  errorCode: z.string().nullable(),
  invoker: z.string(), // TODO: check if this is still true "api", "cron", "cloudevent", "complete" (if it's created as a subflow from another instance it's something like instance:%v, where %v is the instance ID of its parent
  definition: z.string(),
  flow: z.array(z.string()),
  traceId: z.string().nullable(),
  lineage: z.array(z.unknown()), // TODO: this might hold information about its parent instance, refine schema
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
