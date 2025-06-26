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

type DialogState =
  | {
      action: "create";
    }
  | {
      action: "edit";
      index: number;
    }
  | null;

export const Table = <T,>({
  data,
  getItemKey,
  itemLabel,
  label,
  onChange,
  renderForm,
  renderRow,
}: TableProps<T>) => {
  const [dialog, setDialog] = useState<DialogState>(null);
  const [items, setItems] = useState(data);

  const addItem = (item: T) => {
    const newItems = [...items, item];
    setItems(newItems);
    onChange(newItems);
  };

  const updateItem = (targetIndex: number, newItem: T) => {
    const newItems = items.map((item, index) =>
      targetIndex === index ? newItem : item
    );
    setItems(newItems);
    onChange(newItems);
  };

  const moveItem = (srcIndex: number, targetIndex: number) => {
    const newItems = [...items];
    const [targetItem] = newItems.splice(srcIndex, 1);
    if (!targetItem) throw new Error("Invalid source index");
    newItems.splice(targetIndex, 0, targetItem);
    setItems(newItems);
    onChange(newItems);
  };

  const deleteItem = (targetIndex: number) => {
    const newItems = items.filter((_, i) => i !== targetIndex);
    setItems(newItems);
    onChange(newItems);
  };

  const handleSubmit = (item: T) => {
    setDialog(null);
    if (dialog?.action === "edit") {
      updateItem(dialog.index, item);
    } else {
      addItem(item);
    }
  };

  const formValues =
    dialog?.action === "edit" ? items[dialog.index] : undefined;
  const columnCount = items[0] ? renderRow(items[0]).length : 0;

  return (
    <Dialog
      open={dialog !== null}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setDialog(null);
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
                <Button
                  icon
                  variant="outline"
                  size="sm"
                  onClick={() => setDialog({ action: "create" })}
                >
                  <Plus />
                  {itemLabel}
                </Button>
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
                      setDialog({ action: "edit", index });
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
          setDialog(null);
        }}
      >
        {renderForm(formId, handleSubmit, formValues)}
      </ModalWrapper>
    </Dialog>
  );
};
