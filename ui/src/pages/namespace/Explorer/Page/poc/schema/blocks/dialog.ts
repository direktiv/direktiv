import { Block, BlockType, TriggerBlock, TriggerBlockType } from ".";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type DialogType = {
  type: "dialog";
  trigger: TriggerBlockType;
  blocks: BlockType[];
};

export const Dialog = z.object({
  type: z.literal("dialog"),
  trigger: z.lazy(() => TriggerBlock),
  blocks: z.array(z.lazy(() => Block)),
}) satisfies z.ZodType<DialogType>;
