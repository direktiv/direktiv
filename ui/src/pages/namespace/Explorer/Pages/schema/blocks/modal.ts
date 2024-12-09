import { Blocks, BlocksType } from ".";

import { z } from "zod";

export type ModalType = {
  type: "modal";
  data: {
    trigger: BlocksType["trigger"];
    blocks: BlocksType["all"][];
  };
};

export const Modal = z.object({
  type: z.literal("modal"),
  data: z.object({
    trigger: Blocks.trigger,
    blocks: z.array(Blocks.all),
  }),
}) satisfies z.ZodType<ModalType>;
