import { AllBlocks } from "./blocks";
import { z } from "zod";

export const DirektivPagesSchema = z.object({
  direktiv_api: z.literal("pages/v1"),
  blocks: z.array(AllBlocks),
});

export type DirektivPagesType = z.infer<typeof DirektivPagesSchema>;
