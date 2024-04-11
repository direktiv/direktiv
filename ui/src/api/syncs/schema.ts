import { z } from "zod";

const SyncStatusSchema = z.enum(["pending", "executing", "complete", "failed"]);

export const syncSchema = z.object({
  id: z.string(),
  status: SyncStatusSchema,
  endedAt: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const syncListSchema = z.object({
  data: z.array(syncSchema),
});
