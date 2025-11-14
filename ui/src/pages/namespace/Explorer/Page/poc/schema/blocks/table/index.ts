import { BlocklessLoop, BlocklessLoopType } from "../loop";
import { Dialog, DialogType } from "../dialog";
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
  blocks: DialogType[];
};

export const TableActions = z.object({
  type: z.literal("table-actions"),
  blocks: z.array(z.lazy(() => Dialog)),
}) satisfies z.ZodType<TableActionsType>;

export type RowActionsType = {
  type: "row-actions";
  blocks: DialogType[];
};

export const RowActions = z.object({
  type: z.literal("row-actions"),
  blocks: z.array(z.lazy(() => Dialog)),
}) satisfies z.ZodType<RowActionsType>;

export const Table = z.object({
  type: z.literal("table"),
  data: BlocklessLoop,
  blocks: z.tuple([TableActions, RowActions]),
  columns: z.array(TableColumn),
}) satisfies z.ZodType<TableType>;

export type TableType = {
  type: "table";
  data: BlocklessLoopType;
  blocks: [TableActionsType, RowActionsType];
  columns: TableColumnType[];
};
