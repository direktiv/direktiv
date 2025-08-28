import { DatePicker } from "./DatePicker";
import { Fieldset } from "../utils/FieldSet";
import { FormDateInputType } from "../../../../schema/blocks/form/dateInput";
import { encodeBlockKey } from "../utils";
import { useTemplateStringResolver } from "../../../primitives/Variable/utils/useTemplateStringResolver";

type FormDateInputProps = {
  blockProps: FormDateInputType;
};

export const FormDateInput = ({ blockProps }: FormDateInputProps) => {
  const templateStringResolver = useTemplateStringResolver();
  const { id, label, description, defaultValue, optional, type } = blockProps;
  const value = templateStringResolver(defaultValue);
  const fieldName = encodeBlockKey(type, id, optional);
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <DatePicker
        defaultValue={value}
        fieldName={fieldName}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
