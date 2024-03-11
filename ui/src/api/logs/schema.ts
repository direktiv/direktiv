import { z } from "zod";

export const LogLevelSchema = z.enum(["INFO", "ERROR"]);
export type LogLevelSchemaType = z.infer<typeof LogLevelSchema>;

/**
 * example
 * 
{
  "id": "7",
  "time": "2024-02-29T09:22:50.768872Z",
  "msg": "Workflow 'delay.yaml' has been triggered by api.",
  "level": "INFO",
  "trace": "00000000000000000000000000000000",
  "span": "0000000000000000",
  "state": null,
  "branch": null,
  "workflow": "/delay.yaml",
  "instance": "a0f049b0-c04c-4810-8e12-a061ae0f17c5",
  "namespace": "demo",
  "activity": null,
  "route": null,
  "path": null,
  "error": null
}
 */
export const LogEntrySchema = z.object({
  id: z.string().nonempty(),
  time: z.string().nonempty(),
  msg: z.string().nonempty(),
  level: LogLevelSchema,
  trace: z.string().nonempty().nullable(),
  span: z.string().nonempty().nullable(),
  state: z.string().nonempty().nullable(),
  branch: z.string().nonempty().nullable(),
  workflow: z.string().nonempty(),
  instance: z.string().nullable(),
  namespace: z.string().nonempty(),
  activity: z.string().nullable(),
  route: z.string().nullable(),
  path: z.string().nullable(),
  error: z.string().nullable(),
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
  meta: z
    .object({
      previousPage: z.string().nonempty().nullable(),
      startingFrom: z.string().nonempty(),
    })
    .nullable(), // TODO: should not be nullable
  data: z.array(LogEntrySchema).nullable(), // TODO: should not be nullable
});

export type LogsSchemaType = z.infer<typeof LogsSchema>;
export type LogEntryType = z.infer<typeof LogEntrySchema>;
