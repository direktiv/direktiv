import { ArrayItem } from "./ArrayItem";
import { RenderItemType } from "./types";
import { useState } from "react";

type ArrayInputType = <T>(props: {
  defaultValue: T[];
  emptyItem: T;
  onChange: (newValue: T[]) => void;
  itemIsValid?: (item?: T) => boolean;
  renderItem: RenderItemType<T>;
}) => JSX.Element;

export const ArrayInput: ArrayInputType = ({
  defaultValue,
  emptyItem,
  renderItem,
  onChange,
  itemIsValid = () => true,
}) => {
  const [items, setItems] = useState(defaultValue);

  type OneItem = (typeof items)[number];

  const addItem = (newItem: OneItem) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (index: number, value: OneItem) => {
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
    <>
      {items?.map((item, index) => (
        <ArrayItem
          key={`${items.length}-${index}`}
          item={item}
          itemIsValid={itemIsValid}
          renderItem={renderItem}
          onUpdate={(value) => updateAtIndex(index, value)}
          onDelete={() => deleteAtIndex(index)}
        />
      ))}
      <ArrayItem
        item={emptyItem}
        itemIsValid={itemIsValid}
        renderItem={renderItem}
        onAdd={addItem}
      />
    </>
  );
};
