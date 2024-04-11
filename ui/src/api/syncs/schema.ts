import { z } from "zod";

const SyncStatusSchema = z.enum(["pending", "executing", "complete", "failed"]);

export const SyncObjectSchema = z.object({
  id: z.string(),
  status: SyncStatusSchema,
  endedAt: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const SyncResponseSchema = z.object({
  data: SyncObjectSchema,
});

export const syncListSchema = z.object({
  data: z.array(SyncObjectSchema),
});

export type SyncResponseSchemaType = z.infer<typeof SyncResponseSchema>;
export type SyncObjectSchema = z.infer<typeof SyncObjectSchema>;
