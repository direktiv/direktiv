import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";
import { encodeElementKey } from "./utils";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { id, label, description, defaultValue, optional, type } = blockProps;
  const templateStringResolver = useTemplateStringResolver();

  const value = templateStringResolver(defaultValue);
  const fieldName = encodeElementKey(type, id);

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <Textarea
        defaultValue={value}
        id={fieldName}
        name={fieldName}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
