import {
  VariableContextProvider,
  useVariablesContext,
} from "../../primitives/Variable/VariableContext";

import { BlockPathType } from "..";
import { RowActions } from "./actions/RowActions";
import { TableCell } from "./TableCell";
import { TableRow } from "~/design/Table";
import { TableType } from "../../../schema/blocks/table";

export const TableRows = ({
  blocks,
  items,
  total,
  loopId,
  columns,
  blockPath,
}: {
  blocks: TableType["blocks"];
  items: unknown[];
  total: number;
  loopId: TableType["data"]["id"];
  columns: TableType["columns"];
  blockPath: BlockPathType;
}) => {
  const parentVariables = useVariablesContext();

  const hasRowActions = blocks[1].blocks.length > 0;
  const hasTableActions = blocks[0].blocks.length > 0;

  return (
    <>
      {items.map((item, index) => (
        <VariableContextProvider
          key={`${total}-${index}`}
          variables={{
            ...parentVariables,
            loop: {
              ...parentVariables.loop,
              [loopId]: item,
            },
          }}
        >
          <TableRow>
            {columns.map((column, columnIndex) => {
              const isLastColumn = columnIndex === columns.length - 1;
              return (
                <TableCell
                  key={columnIndex}
                  blockProps={column}
                  colSpan={
                    isLastColumn && hasTableActions && !hasRowActions ? 2 : 1
                  }
                />
              );
            })}
            {hasRowActions && (
              <RowActions actions={blocks[1]} blockPath={blockPath} />
            )}
          </TableRow>
        </VariableContextProvider>
      ))}
    </>
  );
};
