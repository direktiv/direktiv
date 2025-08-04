import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { id, label, description, defaultValue } = blockProps;
  return (
    <Fieldset label={label} description={description} htmlFor={id}>
      <Textarea defaultValue={defaultValue} />
    </Fieldset>
  );
};
