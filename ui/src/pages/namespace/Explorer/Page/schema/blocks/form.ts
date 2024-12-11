import { AllBlocks, AllBlocksType, TriggerBlocks, TriggerBlocksType } from ".";
import { Mutation, MutationType } from "../procedures/mutation";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type FormType = {
  type: "form";
  trigger: TriggerBlocksType;
  mutation: MutationType;
  blocks: AllBlocksType[];
};

export const Form = z.object({
  type: z.literal("form"),
  trigger: TriggerBlocks,
  mutation: Mutation,
  blocks: z.array(AllBlocks),
}) satisfies z.ZodType<FormType>;
