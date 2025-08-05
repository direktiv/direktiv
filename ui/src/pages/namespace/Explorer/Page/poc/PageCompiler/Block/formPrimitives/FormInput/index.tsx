import { DateInput } from "./DateInput";
import { Fieldset } from "../utils/FieldSet";
import { FormInputType } from "../../../../schema/blocks/form/input";
import { TextInput } from "./TextInput";

type FormInputProps = {
  blockProps: FormInputType;
};

export const FormInput = ({ blockProps }: FormInputProps) => {
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `id-${id}`;
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
          defaultValue={`${defaultValue}`}
          // remount when defaultValue changes
          key={defaultValue}
        />
      ) : (
        <TextInput
          id={htmlID}
          defaultValue={`${defaultValue}`}
          variant={variant}
        />
      )}
    </Fieldset>
  );
};
