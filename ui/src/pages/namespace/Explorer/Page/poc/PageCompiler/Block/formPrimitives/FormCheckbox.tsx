import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => (
  <Fieldset
    label={blockProps.label}
    description={blockProps.description}
    horizontal
  >
    <Checkbox defaultChecked={blockProps.defaultValue} />
  </Fieldset>
);
