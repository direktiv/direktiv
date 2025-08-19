import { Checkbox as CheckboxDesignComponent } from "~/design/Checkbox";
import { useState } from "react";

type CheckboxProps = {
  defaultValue: boolean;
  fieldName: string;
};

export const Checkbox = ({ defaultValue, fieldName }: CheckboxProps) => {
  const [value, setValue] = useState(defaultValue);

  return (
    <>
      <CheckboxDesignComponent
        checked={value}
        id={fieldName}
        onCheckedChange={(newValue) => {
          if (newValue === "indeterminate") return;
          setValue(newValue);
        }}
      />
      <input type="hidden" name={fieldName} value={String(value)} />
    </>
  );
};
