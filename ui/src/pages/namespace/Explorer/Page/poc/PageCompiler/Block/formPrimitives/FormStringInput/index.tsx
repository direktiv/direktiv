import { DateInput } from "./DateInput";
import { Fieldset } from "../utils/FieldSet";
import { FormStringInputType } from "../../../../schema/blocks/form/stringInput";
import { TextInput } from "./TextInput";
import { useTemplateStringResolver } from "../../../primitives/Variable/utils/useTemplateStringResolver";

type FormInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormInputProps) => {
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
      {value}
      {variant === "date" ? (
        <DateInput
          id={htmlID}
          defaultValue={value}
          // remount when defaultValue changes
          key={value}
        />
      ) : (
        <TextInput
          id={htmlID}
          defaultValue={value}
          variant={variant}
          // remount when defaultValue changes
          key={value}
        />
      )}
    </Fieldset>
  );
};
