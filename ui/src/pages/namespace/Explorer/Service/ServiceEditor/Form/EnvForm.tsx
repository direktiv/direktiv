import { FC, useState } from "react";

import { EnvItemForm } from "./EnvItemForm";
import { EnvironementVariableSchemaType } from "~/api/services/schema/services";

type EnvFormProps = {
  placeholder?: string;
  defaultValue: EnvironementVariableSchemaType[];
  onChange: (newValue: EnvironementVariableSchemaType[]) => void;
};

const EnvForm: FC<EnvFormProps> = ({ defaultValue, onChange }) => {
  const [items, setItems] = useState(defaultValue);

  const addItem = (newItem: EnvironementVariableSchemaType) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (
    index: number,
    value: EnvironementVariableSchemaType
  ) => {
    const newItems = items.map((oldValue, oldIndex) => {
      if (oldIndex === index) {
        return value;
      }
      return oldValue;
    });
    setItems(newItems);
    onChange(newItems);
  };

  return (
    <>
      {items?.map((item, index) => (
        <EnvItemForm
          key={index}
          item={item}
          onUpdate={(value) => updateAtIndex(index, value)}
        />
      ))}
      <EnvItemForm onAdd={addItem} />
    </>
  );
};

export default EnvForm;
