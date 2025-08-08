import { Fieldset } from "./utils/FieldSet";
import { FormNumberInputType } from "../../../schema/blocks/form/numberInput";
import Input from "~/design/Input";

type FormNumberInputProps = {
  blockProps: FormNumberInputType;
};

export const FormNumberInput = ({ blockProps }: FormNumberInputProps) => {
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-input-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Input type="number" defaultValue={defaultValue} id={htmlID} />
    </Fieldset>
  );
};
