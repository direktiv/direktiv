import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { id, label, description, defaultValue, optional } = blockProps;
  const templateStringResolver = useTemplateStringResolver();

  const value = templateStringResolver(defaultValue);
  const htmlID = `form-textarea-${id}`;

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Textarea
        defaultValue={value}
        id={htmlID}
        name={id}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
