import Button from "~/design/Button";
import { RenderItem } from "./types";
import { X } from "lucide-react";

type ArrayItemProps = <T>(props: {
  value: T;
  renderItem: RenderItem<T>;
  onUpdate: (item: T) => void;
  onDelete: () => void;
}) => JSX.Element;

export const ArrayItem: ArrayItemProps = ({
  value,
  renderItem,
  onUpdate,
  onDelete,
}) => {
  type Item = typeof value;

  const setValueAndTriggerCallback = (value: Item) => {
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
