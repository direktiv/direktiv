import { z } from "zod";

export const PodStatusSchema = z.enum([
  "Running",
  "Pending",
  "Succeeded",
  "Failed",
  "Unknown",
]);

/**
 * example
  {
    "name": "namespace-14529307612894023951-00004-deployment-76d465f47cqvfk7",
    "status": "Running",
    "serviceName": "namespace-14529307612894023951",
    "serviceRevision": "namespace-14529307612894023951-00004"
  }
 */

export const PodSchema = z.object({
  name: z.string(),
  status: PodStatusSchema,
  serviceName: z.string(),
  serviceRevision: z.string(),
});

/**
   * example
    {
      "pods": []
    }
   */
export const PodsListSchema = z.object({
  pods: z.array(PodSchema),
});

export const PodsStreamingSchema = z.object({
  event: z.enum(["ADDED", "MODIFIED", "DELETED"]),
  pod: PodSchema,
});

/**
 * example
  {
    "data": "2023/08/18 07:02:13 Serving hello world at http://[::]:8080\n"
  }
 */
export const PodLogsSchema = z.object({
  data: z.string(),
});

export type PodStatusSchemaType = z.infer<typeof PodStatusSchema>;
export type PodSchemaType = z.infer<typeof PodSchema>;
export type PodsListSchemaType = z.infer<typeof PodsListSchema>;
export type PodsStreamingSchemaType = z.infer<typeof PodsStreamingSchema>;
export type PodLogsSchemaType = z.infer<typeof PodLogsSchema>;
