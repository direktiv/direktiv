import { AllBlocks, AllBlocksType } from ".";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type TwoColumnsType = {
  type: "two-columns";
  leftBlocks: AllBlocksType[];
  rightBlocks: AllBlocksType[];
};

export const TwoColumns = z.object({
  type: z.literal("two-columns"),
  leftBlocks: z.array(z.lazy(() => AllBlocks)),
  rightBlocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<TwoColumnsType>;
