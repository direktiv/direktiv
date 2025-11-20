import { Fieldset } from "./utils/FieldSet";
import { FormStringInputType } from "../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { StopPropagation } from "~/components/StopPropagation";
import { encodeBlockKey } from "./utils";
import { useStringInterpolation } from "../../primitives/Variable/utils/useStringInterpolation";

type FormStringInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormStringInputProps) => {
  const { id, label, description, variant, defaultValue, optional, type } =
    blockProps;
  const interpolateString = useStringInterpolation();

  const value = interpolateString(defaultValue);
  const fieldName = encodeBlockKey(type, id, optional);

  return (
    <Fieldset
      id={id}
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <StopPropagation>
        <Input
          id={fieldName}
          name={fieldName}
          defaultValue={value}
          type={variant}
          // remount when defaultValue changes
          key={value}
        />
      </StopPropagation>
    </Fieldset>
  );
};
