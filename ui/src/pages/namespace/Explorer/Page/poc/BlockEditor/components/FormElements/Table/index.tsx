import { Dialog, DialogTrigger } from "~/design/Dialog";
import { ReactNode, useState } from "react";
import {
  TableBody,
  Table as TableDesignComponent,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { ModalWrapper } from "~/components/ModalWrapper";
import { Plus } from "lucide-react";
import { Row } from "./Row";

type TableProps<T> = {
  data: T[];
  getItemKey: (item: T) => string;
  itemLabel: string;
  label: (count: number) => string;
  onChange: (newData: T[]) => void;
  renderRow: (item: T) => string[];
  renderForm: (
    formId: string,
    onSubmit: (item: T) => void,
    defaultValues?: T
  ) => ReactNode;
};

const formId = "table-form-element";

export const Table = <T,>({
  data,
  getItemKey,
  itemLabel,
  label,
  onChange,
  renderForm,
  renderRow,
}: TableProps<T>) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [items, setItems] = useState(data);
  const [editIndex, setEditIndex] = useState<number>();

  const addItem = (item: T) => {
    const newItems = [...items, item];
    setItems(newItems);
    onChange(newItems);
  };

  const updateItem = (targetIndex: number, newItem: T) => {
    const newItems = items.map((item, index) => (targetIndex === index ? newItem : item));
    setItems(newItems);
    onChange(newItems);
  };

  const moveItem = (srcIndex: number, targetIndex: number) => {
    const newItems = [...items];
    const [targetItem] = newItems.splice(srcIndex, 1);
    if (!movedItem) throw new Error("Invalid source index");
    newItems.splice(targetIndex, 0, movedItem);
    setItems(newItems);
    onChange(newItems);
  };

  const deleteItem = (index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    setItems(newItems);
    onChange(newItems);
  };

  const handleSubmit = (item: T) => {
    setDialogOpen(false);
    if (editIndex === undefined) {
      addItem(item);
    } else {
      updateItem(editIndex, item);
    }
    setEditIndex(undefined);
  };

  const formValues = editIndex !== undefined ? items[editIndex] : undefined;
  const columnCount = items[0] ? renderRow(items[0]).length : 0;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <Card noShadow>
        <TableDesignComponent>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell colSpan={columnCount}>
                {label(items.length)}
              </TableHeaderCell>
              <TableHeaderCell className="w-60 text-right">
                <DialogTrigger asChild>
                  <Button icon variant="outline" size="sm">
                    <Plus />
                    {itemLabel}
                  </Button>
                </DialogTrigger>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((item, index, srcArray) => {
              const canMoveDown = index < srcArray.length - 1;
              const canMoveUp = index > 0;

              return (
                <Row
                  key={getItemKey(item)}
                  item={item}
                  renderRow={renderRow}
                  actions={{
                    onEdit: () => {
                      setDialogOpen(true);
                      setEditIndex(index);
                    },
                    onMoveUp: canMoveUp
                      ? () => moveItem(index, index - 1)
                      : undefined,
                    onMoveDown: canMoveDown
                      ? () => moveItem(index, index + 1)
                      : undefined,
                    onDelete: () => deleteItem(index),
                  }}
                />
              );
            })}
          </TableBody>
        </TableDesignComponent>
      </Card>
      <ModalWrapper
        formId={formId}
        title={itemLabel}
        onCancel={() => {
          setDialogOpen(false);
          setEditIndex(undefined);
        }}
      >
        {renderForm(formId, handleSubmit, formValues)}
      </ModalWrapper>
    </Dialog>
  );
};
