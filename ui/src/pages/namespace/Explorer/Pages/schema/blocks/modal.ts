import { Block } from ".";
import { z } from "zod";

export const Modal = z.object({
  type: z.literal("modal"),
  data: z.object({
    trigger: Block.trigger,
    blocks: Block.all,
  }),
});
