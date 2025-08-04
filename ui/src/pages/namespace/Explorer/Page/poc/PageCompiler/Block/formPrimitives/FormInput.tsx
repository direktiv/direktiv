import { Fieldset } from "./utils/FieldSet";
import { FormInputType } from "../../../schema/blocks/form/input";
import Input from "~/design/Input";

type FormInputProps = {
  blockProps: FormInputType;
};

export const FormInput = ({ blockProps }: FormInputProps) => {
  const { id, label, description, variant, defaultValue } = blockProps;
  return (
    <Fieldset label={label} description={description} htmlFor={id}>
      <Input type={variant} defaultValue={defaultValue} />
    </Fieldset>
  );
};
