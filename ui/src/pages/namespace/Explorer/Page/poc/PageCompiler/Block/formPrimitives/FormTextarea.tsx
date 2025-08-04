import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { TemplateString } from "../../primitives/TemplateString";
import { Textarea } from "~/design/TextArea";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => (
  <div>
    <label>
      <TemplateString value={blockProps.label} />
    </label>
    <Textarea defaultValue={blockProps.defaultValue} />
    <p>
      <TemplateString value={blockProps.description} />
    </p>
  </div>
);
