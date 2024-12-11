import { Blocks } from "./blocks";
import { DynamicString } from "./primitives/dynamicString";
import { z } from "zod";

export const DirektivPagesSchema = z.object({
  direktiv_api: z.literal("pages/v1"),
  path: DynamicString,
  blocks: z.array(Blocks.all),
});

// add some basic form input blocks
