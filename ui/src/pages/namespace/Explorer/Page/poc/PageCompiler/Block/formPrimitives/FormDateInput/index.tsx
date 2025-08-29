import { DatePicker } from "./DatePicker";
import { Fieldset } from "../utils/FieldSet";
import { FormDateInputType } from "../../../../schema/blocks/form/dateInput";
import { encodeBlockKey } from "../utils";
import { useStringInterpolation } from "../../../primitives/Variable/utils/useStringInterpolation";

type FormDateInputProps = {
  blockProps: FormDateInputType;
};

export const FormDateInput = ({ blockProps }: FormDateInputProps) => {
  const interpolateString = useStringInterpolation();
  const { id, label, description, defaultValue, optional, type } = blockProps;
  const value = interpolateString(defaultValue);
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
