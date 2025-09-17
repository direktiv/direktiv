import { Block, BlockType } from "..";
import { BlocklessLoop, BlocklessLoopType } from "../loop";
import { TableColumn, TableColumnType } from "./tableColumn";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type TableActionsType = {
  type: "table-actions";
  blocks: BlockType[];
};

export const TableActions = z.object({
  type: z.literal("table-actions"),
  blocks: z.array(Block),
}) satisfies z.ZodType<TableActionsType>;

export const Table = z.object({
  type: z.literal("table"),
  data: BlocklessLoop,
  blocks: z.tuple([TableActions, TableActions]),
  columns: z.array(TableColumn),
}) satisfies z.ZodType<TableType>;

export type TableType = {
  type: "table";
  data: BlocklessLoopType;
  blocks: [TableActionsType, TableActionsType];
  columns: TableColumnType[];
};
