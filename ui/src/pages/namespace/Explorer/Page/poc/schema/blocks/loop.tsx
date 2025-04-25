import { AllBlocks, AllBlocksType } from ".";

import { Variable, VariableType } from "../primitives/variable";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type LoopType = {
  type: "loop";
  id: string;
  variable: VariableType;
  blocks: AllBlocksType[];
};

export const Loop = z.object({
  type: z.literal("loop"),
  id: z.string().min(1),
  variable: Variable,
  blocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<LoopType>;
