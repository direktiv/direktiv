import { z } from "zod";

const instanceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  id: z.string(),
  as: z.string(), // f.e. "some.yaml:latest",
  status: z.string(), // TODO: "complete", what else?
  errorCode: z.string(),
  errorMessage: z.string(),
  invoker: z.string(), // TODO: can be "api", what else?
});

export const InstancesInputSchema = z.object({
  namespace: z.string(),
  instance: instanceSchema,
  data: z.string(),
});
