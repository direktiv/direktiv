import { z } from "zod";

/**
 * example
  {
    "id": "b17bb363d832468bef21_1"
  }
 */

export const PodSchema = z.object({
  id: z.string(),
});

/**
  {
    "data": [
      {
        "id": "b17bb363d832468bef21_1"
      }
    ]
  }
 */
export const PodsListSchema = z.object({
  data: z.array(PodSchema),
});

/**
 * example
  "2023/08/18 07:02:13 Serving hello world at http://[::]:8080\n"
 */
export const PodLogsSchema = z.string();

export type PodSchemaType = z.infer<typeof PodSchema>;
export type PodLogsSchemaType = z.infer<typeof PodLogsSchema>;
