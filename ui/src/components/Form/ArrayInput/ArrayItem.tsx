import { KeyboardEvent, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { RenderItemType } from "./types";

type ArrayItemType = <T>(props: {
  item: T;
  renderItem: RenderItemType<T>;
  itemIsValid: (item?: T) => boolean;
  onAdd?: (value: T) => void;
  onUpdate?: (value: T) => void;
  onDelete?: () => void;
}) => JSX.Element;

export const ArrayItem: ArrayItemType = ({
  item, // TODO: rename to defaultValue
  renderItem,
  itemIsValid,
  onAdd,
  onUpdate,
  onDelete,
}) => {
  type Item = typeof item;
  const [state, setState] = useState<Item>(item);

  const handleChange = (newState: Item) => {
    onUpdate && onUpdate(newState);
  };

  const isValid = itemIsValid(state);

  const handleAdd = () => {
    if (!isValid || !onAdd) return;
    onAdd(state);
    // clear all inputs
    setState(item);
  };

  const handleDelete = () => {
    onDelete && onDelete();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!isValid) return;
    if (event.key === "Enter") {
      event.preventDefault();
      if (!onAdd) return;
      event.currentTarget.blur();
      handleAdd();
    }
  };

  return (
    <ButtonBar data-testid="env-item-form">
      {renderItem({
        state,
        setState,
        onChange: handleChange,
        handleKeyDown,
      })}
      {onAdd && (
        <Button
          type="button"
          variant="outline"
          onClick={handleAdd}
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
