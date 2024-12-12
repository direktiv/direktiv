import { AllBlocks, AllBlocksType, TriggerBlocks, TriggerBlocksType } from ".";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type ModalType = {
  type: "modal";
  trigger: TriggerBlocksType;
  blocks: AllBlocksType[];
};

export const Modal = z.object({
  type: z.literal("modal"),
  trigger: z.lazy(() => TriggerBlocks),
  blocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<ModalType>;
