import { PropsWithChildren, useState } from "react";

import { ArrayItem } from "./ArrayItem";
import Button from "~/design/Button";
import { Plus } from "lucide-react";
import { RenderItem } from "./types";

type ArrayFormProps = <T>(
  props: {
    defaultValue: T[];
    emptyItem: T;
    onChange: (newArray: T[]) => void;
    renderItem: RenderItem<T>;
  } & PropsWithChildren
) => JSX.Element;

export const ArrayForm: ArrayFormProps = ({
  children,
  defaultValue,
  emptyItem,
  onChange,
  renderItem,
}) => {
  type Item = (typeof defaultValue)[number];
  const [items, setItems] = useState(defaultValue);

  const addItem = (newItem: Item) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (index: number, value: Item) => {
    const newItems = items.map((oldValue, oldIndex) => {
      if (oldIndex === index) {
        return value;
      }
      return oldValue;
    });
    setItems(newItems);
    onChange(newItems);
  };

  const deleteAtIndex = (index: number) => {
    const newItems = items.filter((_, oldIndex) => oldIndex !== index);
    setItems(newItems);
    onChange(newItems);
  };

  return (
    <div className="-mx-1 flex max-h-32 flex-col overflow-y-auto p-1">
      {items?.map((item, index) => (
        <ArrayItem
          key={`${items.length}-${index}`}
          defaultValue={item}
          renderItem={renderItem}
          onUpdate={(value) => updateAtIndex(index, value)}
          onDelete={() => deleteAtIndex(index)}
        />
      ))}
      <Button
        type="button"
        variant="outline"
        onClick={(e) => {
          e.preventDefault();
          addItem(emptyItem);
        }}
      >
        <Plus /> {children}
      </Button>
    </div>
  );
};
