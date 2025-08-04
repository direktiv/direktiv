import { Fieldset } from "./utils/FieldSet";
import { FormInputType } from "../../../schema/blocks/form/input";
import Input from "~/design/Input";

type FormInputProps = {
  blockProps: FormInputType;
};

export const FormInput = ({ blockProps }: FormInputProps) => (
  <Fieldset label={blockProps.label} description={blockProps.description}>
    <Input type={blockProps.variant} defaultValue={blockProps.defaultValue} />
  </Fieldset>
);
