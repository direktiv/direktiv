import { z } from "zod";
/**
 * example
  {
    "data": {
      "jx": null,
      "data": "e30=",
      "output": [
        "e30="
      ],
      "logs": ""
    }
  }
 */

export const JxQueryResult = z.object({
  data: z.object({
    jx: z.string().nullable(),
    data: z.string(),
    output: z.array(z.string()),
    logs: z.string(),
  }),
});

export type JxQueryResultType = z.infer<typeof JxQueryResult>;

export const ExecuteJxQueryPayloadSchema = z.object({
  jx: z.string(),
  data: z.string(),
});

export type ExecuteJxQueryPayloadType = z.infer<
  typeof ExecuteJxQueryPayloadSchema
>;
