import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-textarea-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Textarea defaultValue={defaultValue} id={htmlID} />
    </Fieldset>
  );
};
