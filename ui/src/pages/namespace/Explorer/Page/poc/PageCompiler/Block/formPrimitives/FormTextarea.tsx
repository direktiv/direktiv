import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { StopPropagation } from "~/components/StopPropagation";
import { Textarea } from "~/design/TextArea";
import { encodeBlockKey } from "./utils";
import { useStringInterpolation } from "../../primitives/Variable/utils/useStringInterpolation";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { id, label, description, defaultValue, optional, type } = blockProps;
  const interpolateString = useStringInterpolation();

  const value = interpolateString(defaultValue);
  const fieldName = encodeBlockKey(type, id, optional);

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <StopPropagation>
        <Textarea
          defaultValue={value}
          id={fieldName}
          name={fieldName}
          // remount when defaultValue changes
          key={value}
        />
      </StopPropagation>
    </Fieldset>
  );
};
