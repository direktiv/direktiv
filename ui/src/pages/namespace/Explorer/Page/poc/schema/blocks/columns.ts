import { AllBlocks, AllBlocksType } from ".";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type ColumnsType = {
  type: "columns";
  columns: AllBlocksType[][];
};

export const Columns = z.object({
  type: z.literal("columns"),
  columns: z.array(z.array(z.lazy(() => AllBlocks))),
}) satisfies z.ZodType<ColumnsType>;
