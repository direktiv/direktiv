import { DateInput } from "./DateInput";
import { Fieldset } from "../utils/FieldSet";
import { FormStringInputType } from "../../../../schema/blocks/form/stringInput";
import { TextInput } from "./TextInput";

type FormInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormInputProps) => {
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `form-input-${id}`;
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
          defaultValue={String(defaultValue)}
          // remount when defaultValue changes
          key={defaultValue}
        />
      ) : (
        <TextInput
          id={htmlID}
          defaultValue={String(defaultValue)}
          variant={variant}
        />
      )}
    </Fieldset>
  );
};
