import { Fieldset } from "./utils/FieldSet";
import { FormInputType } from "../../../schema/blocks/form/input";
import Input from "~/design/Input";

type FormInputProps = {
  blockProps: FormInputType;
};

export const FormInput = ({ blockProps }: FormInputProps) => {
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `id-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Input type={variant} defaultValue={defaultValue} id={htmlID} />
    </Fieldset>
  );
};
