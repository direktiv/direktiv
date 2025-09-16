import { Block, BlockType, TriggerBlock, TriggerBlockType } from "..";
import { Mutation, MutationType } from "../../procedures/mutation";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type FormType = {
  type: "form";
  trigger: TriggerBlockType;
  mutation: MutationType;
  blocks: BlockType[];
  register?: (fields: string[]) => void;
};

export const Form = z.object({
  type: z.literal("form"),
  trigger: z.lazy(() => TriggerBlock),
  mutation: Mutation,
  blocks: z.array(z.lazy(() => Block)),
}) satisfies z.ZodType<FormType>;
