import { z } from "zod";

const instanceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  id: z.string(),
  as: z.string(), // f.e. "some.yaml:latest",
  status: z.enum(["pending", "failed", "crashed", "complete"]),
  errorCode: z.string(),
  errorMessage: z.string(),
  invoker: z.string(), // "api", "cron", "cloudevent", "complete" (if it's created as a subflow from another instance it's something like instance:%v, where %v is the instance ID of its parent
});

export const InstancesInputSchema = z.object({
  namespace: z.string(),
  instance: instanceSchema,
  data: z.string(),
});
