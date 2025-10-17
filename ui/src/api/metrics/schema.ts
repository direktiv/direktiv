import { z } from "zod";

/**
 * example response
 * 
  {
    "data": {
        "cancelled": 0,
        "complete": 1,
        "running": 0,
        "failed": 2,
        "pending": 0,
        "total": 3
    }
  }
 */

const MetricsObjectSchema = z.object({
  cancelled: z.number(),
  complete: z.number(),
  running: z.number(),
  failed: z.number(),
  pending: z.number(),
  total: z.number(),
});

export const MetricsResponseSchema = z.object({
  data: MetricsObjectSchema,
});

export type MetricsObjectSchemaType = z.infer<typeof MetricsObjectSchema>;
export type MetricsResponseSchemaType = z.infer<typeof MetricsResponseSchema>;
