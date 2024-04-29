import { z } from "zod";

/**
 * example response
 * 
  {
    "data": {
        "cancelled": 0,
        "complete": 1,
        "crashed": 0,
        "failed": 2,
        "pending": 0,
        "total": 3
    }
  }
 */

export const MetricsObjectSchema = z.object({
  cancelled: z.number(),
  complete: z.number(),
  crashed: z.number(),
  failed: z.number(),
  pending: z.number(),
  total: z.number(),
});

export const MetricsResponseSchema = z.object({
  data: MetricsObjectSchema,
});

export type MetricsObjectSchemaType = z.infer<typeof MetricsObjectSchema>;
export type MetricsResponseSchemaType = z.infer<typeof MetricsResponseSchema>;
