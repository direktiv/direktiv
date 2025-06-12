import { AllBlocks, AllBlocksType } from ".";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type ColumnType = {
  type: "column";
  blocks: AllBlocksType[];
};

export const Column = z.object({
  type: z.literal("column"),
  blocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<ColumnType>;

export type ColumnsType = {
  type: "columns";
  blocks: ColumnType[];
};

export const Columns = z.object({
  type: z.literal("columns"),
  blocks: z.array(z.lazy(() => Column)),
}) satisfies z.ZodType<ColumnsType>;
