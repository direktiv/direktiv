import { IsValidItem, RenderItem } from "./types";
import { KeyboardEvent, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";

type ArrayItemProps = <T>(props: {
  defaultValue: T;
  renderItem: RenderItem<T>;
  itemIsValid: IsValidItem<T>;
  onUpdate?: (item: T) => void;
  onAdd?: (item: T) => void;
  onDelete?: () => void;
}) => JSX.Element;

export const ArrayItem: ArrayItemProps = ({
  defaultValue,
  renderItem,
  itemIsValid,
  onAdd,
  onUpdate,
  onDelete,
}) => {
  type Item = typeof defaultValue;

  const [value, setValue] = useState<Item>(defaultValue);
  const isValid = itemIsValid(value);

  const handleChange = (newState: Item) => {
    onUpdate?.(newState);
  };

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

  return (
    <ButtonBar>
      {renderItem({
        value,
        setValue,
        onChange: handleChange,
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
    </ButtonBar>
  );
};
