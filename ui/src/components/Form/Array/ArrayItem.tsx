import Button from "~/design/Button";
import { RenderItem } from "./types";
import { X } from "lucide-react";
import { useState } from "react";

type ArrayItemProps = <T>(props: {
  defaultValue: T;
  renderItem: RenderItem<T>;
  onUpdate: (item: T) => void;
  onDelete: () => void;
}) => JSX.Element;

export const ArrayItem: ArrayItemProps = ({
  defaultValue,
  renderItem,
  onUpdate,
  onDelete,
}) => {
  type Item = typeof defaultValue;

  const [value, setValue] = useState<Item>(defaultValue);

  const setValueAndTriggerCallback = (value: Item) => {
    setValue(value);
    onUpdate(value);
  };

  return (
    <>
      {renderItem({
        value,
        setValue: setValueAndTriggerCallback,
      })}
      <Button
        type="button"
        variant="outline"
        onClick={(e) => {
          e.preventDefault();
          onDelete();
        }}
      >
        <X />
      </Button>
    </>
  );
};
