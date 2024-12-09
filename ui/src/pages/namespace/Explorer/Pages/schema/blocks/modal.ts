import { Blocks, BlocksType } from ".";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
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
