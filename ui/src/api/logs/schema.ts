import { z } from "zod";

/**
 * example
 * 
{
  "id": "164",
  "time": "2024-02-27T08:57:45.421088Z",
  "msg": "Starting workflow",
  "level": "INFO",
  "workflow": "/child.yaml",
  "namespace": "demo",
  "trace": "00000000000000000000000000000000", // only for instances?
  "span": "0000000000000000", // only for instances?
}
 */
export const LogEntrySchema = z.object({
  id: z.string().nonempty(),
  time: z.string().nonempty(),
  msg: z.string().nonempty(),
  level: z.string().nonempty(), // TODO: use schema for log level
  workflow: z.string().nonempty(),
  namespace: z.string().nonempty(),
  trace: z.string().nonempty().optional(),
  span: z.string().nonempty().optional(),
});

/**
 * example
 * 
  {
    "nextPage": "2024-01-17T01:44:08.128136Z",
    "data": []
  }
 */
export const LogsSchema = z.object({
  nextPage: z.string().nonempty().nullable(),
  data: z.array(LogEntrySchema),
});

export type LogsSchemaType = z.infer<typeof LogsSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
