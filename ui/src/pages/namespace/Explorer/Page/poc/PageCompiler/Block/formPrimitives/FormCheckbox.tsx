import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-checkbox-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      horizontal
      optional={optional}
    >
      <Checkbox defaultChecked={defaultValue} id={htmlID} />
    </Fieldset>
  );
};
