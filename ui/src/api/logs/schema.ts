import { z } from "zod";

export const LogLevelSchema = z.enum(["INFO", "ERROR", "WARN", "DEBUG"]);
export type LogLevelSchemaType = z.infer<typeof LogLevelSchema>;

/**
 * example
 * 
{
  "status": "completed",
  "state": "loop",
  "branch": null,
  "workflow": "/wf.yaml",
  "calledAs": null,
  "instance": "f1242c40-3cd9-48bb-82aa-02275df6e1da" 
}
 */
export const WorkflowStatusData = z.object({
  status: z.string().nonempty(),
  state: z.string().nonempty(),
  branch: z.number().nullable(),
  workflow: z.string().nonempty(),
  calledAs: z.string().nullable(),
  instance: z.string().nonempty(),
});

/**
 * example
 * 
{
  "id": 1731,
  "time": "2024-03-11T13:39:13.214148Z",
  "msg": "Running state logic",
  "level": "INFO",
  "namespace": "test",
  "trace": "00000000000000000000000000000000",
  "span": "0000000000000000",
  "workflow": {...},
  "error": null
}
 */
export const LogEntrySchema = z.object({
  id: z.number(),
  time: z.string().nonempty(),
  msg: z.string().nonempty(),
  level: LogLevelSchema,
  namespace: z.string().nonempty().nullable(),
  trace: z.string().nonempty().nullable(),
  span: z.string().nonempty().nullable(),
  error: z.string().nullable(),
  workflow: WorkflowStatusData.optional(),
});

/**
 * example
 * 
  {
    "meta": {
      "previousPage": null,
      "startingFrom": "2024-03-11T13:35:33.318740761Z"
    },
    "data": []
  }
 */
export const LogsSchema = z.object({
  meta: z.object({
    previousPage: z.string().nonempty().nullable(),
    startingFrom: z.string().nonempty().nullable(),
  }),
  data: z.array(LogEntrySchema),
});

export type LogsSchemaType = z.infer<typeof LogsSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
