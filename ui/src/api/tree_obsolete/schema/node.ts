import { z } from "zod";

export const WorkflowStartedSchema = z.object({
  namespace: z.string(),
  instance: z.string(),
});

export const fileNameSchema = z
  .string()
  .regex(/^(([a-z][a-z0-9_\-.]*[a-z0-9])|([a-z]))$/, {
    message:
      "Please use a name that only contains lowercase letters, use - or _ instead of whitespaces.",
  });
