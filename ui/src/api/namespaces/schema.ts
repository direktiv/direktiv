import { LogLevelSchema } from "../schema";
import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
});

export const NamespaceListSchema = z.object({
  results: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  namespace: NamespaceSchema,
});

/**
 * Example log entry
 * 
 * {
      "t":  "2023-08-14T08:22:00.692787Z",
      "level":  "info",
      "msg":  "Workflow /delay.yaml completed.",
      "tags":  {
        "callpath":  "/9ab6abab-23b1-4c8c-9ad0-53a70d0d2c47",
        "instance-id":  "9ab6abab-23b1-4c8c-9ad0-53a70d0d2c47",
        "invoker":  "api",
        "level":  "info",
        "log_instance_call_path":  "/9ab6abab-23b1-4c8c-9ad0-53a70d0d2c47",
        "namespace":  "stefan",
        "namespace-id":  "c75454f2-3790-4f36-a1a2-22ca8a4f8020",
        "recipientType":  "namespace",
        "revision-id":  "908be548-ec50-4a43-94dd-2da717159685",
        "root-instance-id":  "9ab6abab-23b1-4c8c-9ad0-53a70d0d2c47",
        "root_instance_id":  "9ab6abab-23b1-4c8c-9ad0-53a70d0d2c47",
        "source":  "c75454f2-3790-4f36-a1a2-22ca8a4f8020",
        "trace":  "00000000000000000000000000000000",
        "type":  "namespace",
        "workflow":  "/delay.yaml"
      }
    },
 */
export const NamespaceLogSchema = z.object({
  t: z.string(), // 2023-08-07T08:09:49.406596Z
  level: LogLevelSchema,
  msg: z.string(), // Starting workflow /stable-diffusion.yaml
});

export const NamespaceLogListSchema = z.object({
  results: z.array(NamespaceLogSchema),
});

export const NamespaceDeletedSchema = z.null();

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
export type NamespaceLogListSchemaType = z.infer<typeof NamespaceLogListSchema>;
export type NamespaceLogSchemaType = z.infer<typeof NamespaceLogSchema>;
