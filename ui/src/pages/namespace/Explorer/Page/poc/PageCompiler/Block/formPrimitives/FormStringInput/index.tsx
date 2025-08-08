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
  const resolvedDefaultValue = templateStringResolver(defaultValue);

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      {variant === "date" ? (
        <DateInput
          id={htmlID}
          defaultValue={resolvedDefaultValue}
          // remount when defaultValue changes
          key={defaultValue}
        />
      ) : (
        <TextInput
          id={htmlID}
          defaultValue={resolvedDefaultValue}
          variant={variant}
        />
      )}
    </Fieldset>
  );
};
