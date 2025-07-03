import { TableCell, TableRow } from "~/design/Table";

import { ListContextMenu } from "~/components/ListContextMenu";

type RowActions = {
  onEdit: () => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
  onDelete: () => void;
};

type RowProps<T> = {
  item: T;
  renderRow: (item: T) => string[];
  numberOfColumns: number;
  actions: RowActions;
};

export const Row = <T,>({
  item,
  renderRow,
  numberOfColumns,
  actions,
}: RowProps<T>) => {
  const rowCells = renderRow(item);
  return (
    <TableRow
      className="cursor-pointer hover:underline"
      onClick={actions.onEdit}
    >
      {rowCells.map((cell, cellIndex) => (
        <TableCell key={cellIndex} title={cell}>
          <div className="truncate" style={{ width: 200 / numberOfColumns }}>
            {cell} {numberOfColumns}
          </div>
        </TableCell>
      ))}
      <TableCell className="w-0 text-right">
        <ListContextMenu
          onDelete={actions.onDelete}
          onMoveDown={actions.onMoveDown}
          onMoveUp={actions.onMoveUp}
        />
      </TableCell>
    </TableRow>
  );
};
