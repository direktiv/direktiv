import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const { id, label, description, defaultValue } = blockProps;
  return (
    <Fieldset label={label} description={description} htmlFor={id} horizontal>
      <Checkbox defaultChecked={defaultValue} />
    </Fieldset>
  );
};
