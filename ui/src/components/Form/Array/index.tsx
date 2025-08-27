import { IsValidItem, RenderItem } from "./types";
import { ReactNode, useState } from "react";

import { ArrayItem } from "./ArrayItem";

type ArrayFormProps = <T>(props: {
  defaultValue: T[];
  emptyItem: T;
  onChange: (newArray: T[]) => void;
  itemIsValid: IsValidItem<T>;
  renderItem: RenderItem<T>;
  wrapItem?: (children: ReactNode) => JSX.Element;
}) => JSX.Element;

export const ArrayForm: ArrayFormProps = ({
  defaultValue,
  emptyItem,
  renderItem,
  onChange,
  itemIsValid = () => true,
  wrapItem,
}) => {
  const [items, setItems] = useState(defaultValue);

  type Item = (typeof items)[number];

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
    <>
      {items?.map((item, index) => (
        <ArrayItem
          key={`${items.length}-${index}`}
          defaultValue={item}
          itemIsValid={itemIsValid}
          renderItem={renderItem}
          onUpdate={(value) => updateAtIndex(index, value)}
          onDelete={() => deleteAtIndex(index)}
          wrapItem={wrapItem}
        />
      ))}
      <ArrayItem
        defaultValue={emptyItem}
        itemIsValid={itemIsValid}
        renderItem={renderItem}
        onAdd={addItem}
        wrapItem={wrapItem}
      />
    </>
  );
};
