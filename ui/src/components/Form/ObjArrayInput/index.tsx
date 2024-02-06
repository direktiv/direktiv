import { FC, useState } from "react";

import { EnvItemForm } from "./EnvItemForm";
import { UnknownObjectOfStrings } from "./types";

type ObjArrayInputProps = {
  placeholder?: string;
  defaultValue: UnknownObjectOfStrings[];
  onChange: (newValue: UnknownObjectOfStrings[]) => void;
};

const ObjArrayInput: FC<ObjArrayInputProps> = ({ defaultValue, onChange }) => {
  const [items, setItems] = useState(defaultValue);

  const addItem = (newItem: UnknownObjectOfStrings) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (index: number, value: UnknownObjectOfStrings) => {
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
        <EnvItemForm
          key={`${items.length}-${index}`}
          item={item}
          onUpdate={(value) => updateAtIndex(index, value)}
          onDelete={() => deleteAtIndex(index)}
        />
      ))}
      <EnvItemForm onAdd={addItem} />
    </>
  );
};

export default ObjArrayInput;
