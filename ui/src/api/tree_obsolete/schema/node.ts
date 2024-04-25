import { z } from "zod";

export const WorkflowStartedSchema = z.object({
  namespace: z.string(),
  instance: z.string(),
});
