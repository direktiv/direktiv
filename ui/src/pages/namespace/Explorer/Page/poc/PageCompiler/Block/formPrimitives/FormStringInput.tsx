import { Fieldset } from "./utils/FieldSet";
import { FormStringInputType } from "../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";

type FormStringInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormStringInputProps) => {
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `form-input-${id}`;

  const templateStringResolver = useTemplateStringResolver();
  const value = templateStringResolver(defaultValue);

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Input
        id={htmlID}
        name={id}
        defaultValue={value}
        type={variant}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
