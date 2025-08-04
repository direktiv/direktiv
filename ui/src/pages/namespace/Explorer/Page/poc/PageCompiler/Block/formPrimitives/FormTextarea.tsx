import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => (
  <Fieldset label={blockProps.label} description={blockProps.description}>
    <Textarea defaultValue={blockProps.defaultValue} />
  </Fieldset>
);
