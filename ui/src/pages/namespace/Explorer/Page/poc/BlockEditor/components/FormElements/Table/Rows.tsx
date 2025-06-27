import { Row } from "./Row";

type RowsProps<T> = {
  items: T[];
  renderRow: (item: T) => string[];
  onEdit: (targetIndex: number) => void;
  onChange: (newData: T[]) => void;
};

export const Rows = <T,>({
  items,
  renderRow,
  onEdit,
  onChange,
}: RowsProps<T>) => {
  const moveItem = (srcIndex: number, targetIndex: number) => {
    const newItems = [...items];
    const [targetItem] = newItems.splice(srcIndex, 1);
    if (!targetItem) throw new Error("Invalid source index");
    newItems.splice(targetIndex, 0, targetItem);
    onChange(newItems);
  };

  const deleteItem = (targetIndex: number) => {
    const newItems = items.filter((_, i) => i !== targetIndex);
    onChange(newItems);
  };

  return items.map((item, index, srcArray) => {
    const canMoveDown = index < srcArray.length - 1;
    const canMoveUp = index > 0;
    return (
      <Row
        key={index}
        item={item}
        renderRow={renderRow}
        actions={{
          onEdit: () => {
            onEdit(index);
          },
          onMoveUp: canMoveUp ? () => moveItem(index, index - 1) : undefined,
          onMoveDown: canMoveDown
            ? () => moveItem(index, index + 1)
            : undefined,
          onDelete: () => deleteItem(index),
        }}
      />
    );
  });
};
