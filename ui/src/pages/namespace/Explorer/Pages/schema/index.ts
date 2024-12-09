import { Block } from "./blocks";
import { z } from "zod";

export const DirektivPagesSchema = z.object({
  direktiv_api: z.literal("ui/v1"),
  path: z.string().min(1),
  blocks: Block.all,
});
