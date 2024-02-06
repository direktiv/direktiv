import { KeyboardEvent, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { RenderItemType } from "./types";

type ObjArrayInputType = <T extends Readonly<unknown>>(props: {
  item: T;
  renderItem: RenderItemType<T>;
  itemIsValid: (item?: T) => boolean;
  onAdd?: (value: T) => void;
  onUpdate?: (value: T) => void;
  onDelete?: () => void;
}) => JSX.Element;

export const ItemForm: ObjArrayInputType = ({
  item,
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
    if (!onAdd || !state) return;
    onAdd(state);
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
      // set back to default value
      setState(item);
    }
  };

  return (
    <div className="flex gap-3" data-testid="env-item-form">
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
    </div>
  );
};
