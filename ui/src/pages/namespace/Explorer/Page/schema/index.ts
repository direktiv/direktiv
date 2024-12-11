import { AllBlocks } from "./blocks";
import { DynamicString } from "./primitives/dynamicString";
import { z } from "zod";

export const DirektivPagesSchema = z.object({
  direktiv_api: z.literal("pages/v1"),
  path: DynamicString,
  blocks: z.array(AllBlocks),
});

export type DirektivPagesType = z.infer<typeof DirektivPagesSchema>;
