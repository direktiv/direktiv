import { TableCell, TableRow } from "~/design/Table";

import { ListContextMenu } from "~/components/ListContextMenu";
import { ReactNode } from "react";

export type RowActions = {
  onEdit: () => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
  onDelete: () => void;
};

export type RowProps<T> = {
  item: T;
  renderRow: (item: T) => ReactNode[];
  actions: RowActions;
};

export const Row = <T,>({ item, renderRow, actions }: RowProps<T>) => {
  const rowCells = renderRow(item);
  return (
    <TableRow className="cursor-pointer" onClick={actions.onEdit}>
      {rowCells.map((cell, cellIndex) => (
        <TableCell key={cellIndex}>{cell}</TableCell>
      ))}
      <TableCell className="text-right">
        <ListContextMenu
          onDelete={actions.onDelete}
          onMoveDown={actions.onMoveDown}
          onMoveUp={actions.onMoveUp}
        />
      </TableCell>
    </TableRow>
  );
};
