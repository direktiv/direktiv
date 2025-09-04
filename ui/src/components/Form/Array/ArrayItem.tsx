import { IsValidItem, RenderItem } from "./types";
import { KeyboardEvent, ReactNode, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";

type ArrayItemProps = <T>(
  props: {
    defaultValue: T;
    renderItem: RenderItem<T>;
    itemIsValid: IsValidItem<T>;
    wrapItem?: (children: ReactNode) => JSX.Element;
  } & ( // either allow onAdd or onUpdate and onDelete
    | {
        onAdd: (item: T) => void;
        onUpdate?: never;
        onDelete?: never;
      }
    | {
        onAdd?: never;
        onUpdate: (item: T) => void;
        onDelete: () => void;
      }
  )
) => JSX.Element;

export const ArrayItem: ArrayItemProps = ({
  defaultValue,
  renderItem,
  itemIsValid,
  onAdd,
  onUpdate,
  onDelete,
  wrapItem = (children) => <ButtonBar>{children}</ButtonBar>,
}) => {
  type Item = typeof defaultValue;

  const [value, setValue] = useState<Item>(defaultValue);
  const isValid = itemIsValid(value);

  const handleAdd = () => {
    if (!isValid || !onAdd) return;
    onAdd(value);
    setValue(defaultValue);
  };

  const handleDelete = () => {
    onDelete?.();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      // make sure not accidentally submitting a form
      event.preventDefault();
      if (!onAdd || !isValid) return;
      handleAdd();
    }
  };

  const setValueAndTriggerCallback = (value: Item) => {
    setValue(value);
    onUpdate?.(value);
  };

  return wrapItem(
    <>
      {renderItem({
        value,
        setValue: setValueAndTriggerCallback,
        handleKeyDown,
      })}
      {onAdd && (
        <Button
          type="button"
          variant="outline"
          onClick={() => handleAdd()}
          disabled={!isValid}
        >
          <Plus />
        </Button>
      )}
      {onDelete && (
        <Button type="button" variant="outline" onClick={() => handleDelete()}>
          <X />
        </Button>
      )}
    </>
  );
};
