import { Block, BlockType } from ".";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type CardType = {
  type: "card";
  blocks: BlockType[];
};

export const Card = z.object({
  type: z.literal("card"),
  blocks: z.array(z.lazy(() => Block)),
}) satisfies z.ZodType<CardType>;
