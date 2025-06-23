import { Dialog, DialogTrigger } from "~/design/Dialog";
import { ReactNode, useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { ListContextMenu } from "~/components/ListContextMenu";
import { ModalWrapper } from "~/components/ModalWrapper";
import { Plus } from "lucide-react";

export type GenericTableProps<T> = {
  data: T[];
  onChange: (newData: T[]) => void;
  label: string;
  renderRow: (item: T, index: number) => ReactNode[];
  getItemKey: (item: T, index: number) => string;
  renderForm: (
    formId: string,
    onSubmit: (item: T) => void,
    defaultValues?: T
  ) => ReactNode;
};

const formId = "generic-table-form";

export const GenericTable = <T,>({
  data,
  onChange,
  label,
  renderRow,
  getItemKey,
  renderForm,
}: GenericTableProps<T>) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [items, setItems] = useState(data);
  const [editIndex, setEditIndex] = useState<number>();

  const addItem = (item: T) => {
    const newItems = [...items, item];
    setItems(newItems);
    onChange(newItems);
  };

  const updateItem = (index: number, item: T) => {
    const newItems = items.map((i, idx) => (idx === index ? item : i));
    setItems(newItems);
    onChange(newItems);
  };

  const moveItem = (srcIndex: number, targetIndex: number) => {
    const newItems = [...items];
    const [movedItem] = newItems.splice(srcIndex, 1);
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

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <Card noShadow>
        <Table>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell className="w-60 text-right" colSpan={3}>
                <DialogTrigger asChild>
                  <Button icon variant="outline" size="sm">
                    <Plus />
                    {label}
                  </Button>
                </DialogTrigger>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((item, index, srcArray) => {
              const canMoveDown = index < srcArray.length - 1;
              const canMoveUp = index > 0;
              const onMoveUp = canMoveUp
                ? () => {
                    moveItem(index, index - 1);
                  }
                : undefined;
              const onMoveDown = canMoveDown
                ? () => {
                    moveItem(index, index + 1);
                  }
                : undefined;

              const onDelete = () => {
                deleteItem(index);
              };

              const rowCells = renderRow(item, index);

              return (
                <TableRow
                  key={getItemKey(item, index)}
                  className="cursor-pointer"
                  onClick={() => {
                    setDialogOpen(true);
                    setEditIndex(index);
                  }}
                >
                  {rowCells.map((cell, cellIndex) => (
                    <TableCell key={cellIndex}>{cell}</TableCell>
                  ))}
                  <TableCell className="text-right">
                    <ListContextMenu
                      onDelete={onDelete}
                      onMoveDown={onMoveDown}
                      onMoveUp={onMoveUp}
                    />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Card>
      <ModalWrapper
        formId={formId}
        title={label}
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
