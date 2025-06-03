import { z } from "zod";

const LogLevelSchema = z.enum(["INFO", "ERROR", "WARN", "DEBUG"]);
export type LogLevelSchemaType = z.infer<typeof LogLevelSchema>;

/**
 * example
 * 
{
  "status": "running",
  "state": "input",
  "workflow": "/error.yaml",
  "instance": "64477403-24a6-481b-a871-9fb8bc19e648",
  "callpath": "/454c7a14-77d5-4706-bc61-27d883a2dde0/64477403-24a6-481b-a871-9fb8bc19e648/"
}
 */
const WorkflowStatusData = z.object({
  status: z.string().nonempty().nullable(),
  state: z.string().nonempty().nullable(),
  workflow: z.string().nonempty().optional(),
  instance: z.string().nonempty().optional(),
  callpath: z.string().nullable().optional(),
});

/**
 * example
 * 
{
  "path": "/mypath" 
}
 */
const RouteData = z.object({
  path: z.string().nonempty(),
});

/**
 * example
 * 
{
  "time": "2025-03-24T08:30:38.782702844Z",
  "msg": "running state logic prep",
  "level": "INFO",
  "namespace": "demo",
}
 */
export const LogEntrySchema = z.object({
  time: z.string().nonempty(),
  msg: z.string().nonempty(),
  level: LogLevelSchema,
  namespace: z.string().nonempty().nullable(),
  workflow: WorkflowStatusData.optional(),
  route: RouteData.optional(),
});

/**
 * example
 * 
  {
    "data": []
  }
 */
export const LogsSchema = z.object({
  data: z.array(LogEntrySchema),
});

export type LogsSchemaType = z.infer<typeof LogsSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
