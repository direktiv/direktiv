import { AllBlocks } from "./blocks";
import { z } from "zod";

export const DirektivPagesSchema = z.object({
  direktiv_api: z.literal("page/v1"),
  type: z.literal("page"),
  blocks: z.array(AllBlocks),
});

export type DirektivPagesType = z.infer<typeof DirektivPagesSchema>;
