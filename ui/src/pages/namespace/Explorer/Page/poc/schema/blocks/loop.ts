import { Block, BlockType } from ".";

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
  data: VariableType;
  blocks: BlockType[];
  pageSize: number;
};

export const Loop = z.object({
  type: z.literal("loop"),
  id: z.string().min(1),
  data: Variable,
  blocks: z.array(z.lazy(() => Block)),
  pageSize: z.number().min(1),
}) satisfies z.ZodType<LoopType>;

export const BlocklessLoop = Loop.omit({ blocks: true });
export type BlocklessLoopType = z.infer<typeof BlocklessLoop>;
