import { Fieldset } from "./utils/FieldSet";
import { FormStringInputType } from "../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { StopPropagation } from "~/components/StopPropagation";
import { encodeElementKey } from "./utils";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";

type FormStringInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormStringInputProps) => {
  const { id, label, description, variant, defaultValue, optional, type } =
    blockProps;
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
