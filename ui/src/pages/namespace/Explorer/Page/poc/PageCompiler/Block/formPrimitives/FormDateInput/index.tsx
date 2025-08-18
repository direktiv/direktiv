import { DatePicker } from "./DatePicker";
import { Fieldset } from "../utils/FieldSet";
import { FormDateInputType } from "../../../../schema/blocks/form/dateInput";
import { useTemplateStringResolver } from "../../../primitives/Variable/utils/useTemplateStringResolver";

type FormDateInputProps = {
  blockProps: FormDateInputType;
};

export const FormDateInput = ({ blockProps }: FormDateInputProps) => {
  const templateStringResolver = useTemplateStringResolver();
  const { id, label, description, defaultValue, optional } = blockProps;

  const value = templateStringResolver(defaultValue);
  const htmlID = `form-input-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <DatePicker
        defaultValue={value}
        id={htmlID}
        name={id}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
