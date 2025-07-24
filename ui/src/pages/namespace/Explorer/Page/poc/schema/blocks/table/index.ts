import { BlocklessLoop, BlocklessLoopType } from "../loop";
import { TableColumn, TableColumnType } from "./tableColumn";
import { TriggerBlock, TriggerBlockType } from "..";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type TableType = {
  type: "table";
  data: BlocklessLoopType;
  actions: TriggerBlockType[];
  columns: TableColumnType[];
};

export const Table = z.object({
  type: z.literal("table"),
  data: BlocklessLoop,
  actions: z.array(z.lazy(() => TriggerBlock)),
  columns: z.array(TableColumn),
}) satisfies z.ZodType<TableType>;
