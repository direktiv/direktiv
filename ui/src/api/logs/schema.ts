import { z } from "zod";

/**
 * example
 * 
{
  "id": "27219",
  "time": "2024-02-28T08:13:08.369703Z",
  "msg": "Workflow completed.",
  "level": "INFO",
  "trace": "00000000000000000000000000000000", // only for instances?
  "span": "0000000000000000",
  "state": "",
  "branch": "",
  "workflow": "/loop.yaml",
  "instance": "",
  "namespace": "demo",
  "activity": "",
  "route": "",
  "path": "",
  "error": ""
}
 */
export const LogEntrySchema = z.object({
  id: z.string().nonempty(),
  time: z.string().nonempty(),
  msg: z.string().nonempty(),
  level: z.string().nonempty(), // TODO: use schema for log level
  trace: z.string(),
  span: z.string(),
  state: z.string(),
  branch: z.string(),
  workflow: z.string().nonempty(),
  instance: z.string(),
  namespace: z.string().nonempty(),
  activity: z.string(),
  route: z.string(),
  path: z.string(),
  error: z.string(),
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
  next_page: z.string(), // TODO:  must be z.string().nonempty().nullable() and must be changed to next_page
  data: z.array(LogEntrySchema),
});

export type LogsSchemaType = z.infer<typeof LogsSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
