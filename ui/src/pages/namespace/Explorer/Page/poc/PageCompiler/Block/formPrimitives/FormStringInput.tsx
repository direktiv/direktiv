import { Fieldset } from "./utils/FieldSet";
import { FormStringInputType } from "../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";

type FormStringInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormStringInputProps) => {
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `form-input-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Input id={id} defaultValue={defaultValue} type={variant} />
    </Fieldset>
  );
};
