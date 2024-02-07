import { FC, useState } from "react";

import { EnvItemForm } from "./EnvItemForm";
import { EnvVarSchemaType } from "~/api/services/schema/services";

type EnvFormProps = {
  placeholder?: string;
  defaultValue: EnvVarSchemaType[];
  onChange: (newValue: EnvVarSchemaType[]) => void;
};

const EnvForm: FC<EnvFormProps> = ({ defaultValue, onChange }) => {
  const [items, setItems] = useState(defaultValue);

  const addItem = (newItem: EnvVarSchemaType) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (index: number, value: EnvVarSchemaType) => {
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

export default EnvForm;
