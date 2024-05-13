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

export const JqQueryResult = z.object({
  data: z.object({
    jx: z.string().nullable(),
    data: z.string(),
    output: z.array(z.string()),
    logs: z.string(),
  }),
});

export type JqQueryResultType = z.infer<typeof JqQueryResult>;

export const ExecuteJqueryPayloadSchema = z.object({
  jx: z.string(),
  data: z.string(),
});

export type ExecuteJqueryPayloadType = z.infer<
  typeof ExecuteJqueryPayloadSchema
>;
