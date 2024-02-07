import { IsValidItem, RenderItem } from "./types";

import { ArrayItem } from "./ArrayItem";
import { useState } from "react";

type ArrayInputProps = <T>(props: {
  defaultValue: T[];
  emptyItem: T;
  onChange: (newArray: T[]) => void;
  itemIsValid: IsValidItem<T>;
  renderItem: RenderItem<T>;
}) => JSX.Element;

export const ArrayInput: ArrayInputProps = ({
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
