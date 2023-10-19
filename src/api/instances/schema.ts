import { PageinfoSchema } from "../schema";
import { z } from "zod";

const InstanceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  id: z.string(),
  as: z.string(), // f.e. "some.yaml",
  status: z.enum(["pending", "failed", "crashed", "complete"]),
  errorCode: z.string(),
  errorMessage: z.string(),
  invoker: z.string(), // "api", "cron", "cloudevent", "complete" (if it's created as a subflow from another instance it's something like instance:%v, where %v is the instance ID of its parent
});

export const InstancesListSchema = z.object({
  namespace: z.string(),
  instances: z.object({
    pageInfo: PageinfoSchema,
    results: z.array(InstanceSchema),
  }),
});

export const InstancesDetailSchema = z.object({
  namespace: z.string(),
  instance: InstanceSchema,
  invokedBy: z.string(), // mostly empty
  flow: z.array(z.string()), // required for the diagram: a list of states that have been executed
});

export const InstancesInputSchema = z.object({
  namespace: z.string(),
  instance: InstanceSchema,
  data: z.string(),
});

export const InstancesOutputSchema = z.object({
  namespace: z.string(),
  instance: InstanceSchema,
  data: z.string(),
});

export type InstanceSchemaType = z.infer<typeof InstanceSchema>;
export type InstancesDetailSchemaType = z.infer<typeof InstancesDetailSchema>;
