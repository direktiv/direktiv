import { Fieldset } from "./utils/FieldSet";
import { FormStringInputType } from "../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { serializeFieldName } from "./utils";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";

type FormStringInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormStringInputProps) => {
  const { id, label, description, variant, defaultValue, optional, type } =
    blockProps;
  const templateStringResolver = useTemplateStringResolver();

  const value = templateStringResolver(defaultValue);
  const fieldName = serializeFieldName(type, id);

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <Input
        id={fieldName}
        name={fieldName}
        defaultValue={value}
        type={variant}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
