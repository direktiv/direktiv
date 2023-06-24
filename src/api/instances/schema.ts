import { PageinfoSchema } from "../schema";
import { z } from "zod";

const InstanceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  id: z.string(),
  as: z.string(), // f.e. "some.yaml:latest",
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

export const InstancesInputSchema = z.object({
  namespace: z.string(),
  instance: InstanceSchema,
  data: z.string(),
});
