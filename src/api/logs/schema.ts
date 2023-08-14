import { z } from "zod";

const LogEntrySchema = z.object({
  t: z.string(), // "2023-07-27T09:49:58.408869Z"
  level: z.enum(["debug", "info", "error", "panic"]), // "debug" | "info" | "error" | "panic";
  msg: z.string(), // "Preparing workflow triggered by api."
  tags: z.object({
    callpath: z.string(), // "/"
    "instance-id": z.string(), // "4b833cfd-a0ef-4f4c-994b-4b63c7301778"
    invoker: z.string(), // "api", "cron", "cloudevent", "complete" (if it's created as a subflow from another instance it's something like instance:%v, where %v is the instance ID of its parent
    "loop-index": z.string().optional(), // ""
    namespace: z.string(), // "workflow"
    "namespace-id": z.string(), // "d1df8d21-96ce-49eb-98b7-5af7571ab28f"
    recipientType: z.string(), // "instance"
    workflow: z.string(), // "workflow.yaml"
    "workflow-id": z.string().optional(), // "a3acd2eb-75de-4369-bb1e-41d5364ea6b7"
    "state-id": z.string().optional(), // "a3acd2eb-75de-4369-bb1e-41d5364ea6b7"
  }),
});

export const LogListSchema = z.object({
  namespace: z.string(),
  instance: z.string(),
  results: z.array(LogEntrySchema),
});

export type LogListSchemaType = z.infer<typeof LogListSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
