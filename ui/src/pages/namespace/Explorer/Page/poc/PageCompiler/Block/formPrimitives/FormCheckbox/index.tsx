import { Checkbox } from "./Checkbox";
import { Fieldset } from "../utils/FieldSet";
import { FormCheckboxType } from "../../../../schema/blocks/form/checkbox";
import { StopPropagation } from "~/components/StopPropagation";
import { encodeBlockKey } from "../utils";
import { usePageStateContext } from "../../../context/pageCompilerContext";
import { useUnwrapOrThrow } from "../../../primitives/Variable/utils/useUnwrapOrThrow";
import { useVariableBooleanResolver } from "../../../primitives/Variable/utils/useVariableBooleanResolver";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const unwrapOrThrow = useUnwrapOrThrow();
  const resolveVariableBoolean = useVariableBooleanResolver();
  const { id, label, description, defaultValue, optional, type } = blockProps;
  const { mode } = usePageStateContext();

  const fieldName = encodeBlockKey(type, id, optional);
  let value: boolean;

  if (defaultValue.type === "variable") {
    const resolvedDefaultValue = resolveVariableBoolean(defaultValue.value);
    value = unwrapOrThrow(resolvedDefaultValue, defaultValue.value);
  } else {
    value = defaultValue.value;
  }

  return (
    <Fieldset
      id={id}
      label={label}
      description={description}
      htmlFor={fieldName}
      horizontal
      optional={optional}
      onClickLabel={(event) => mode === "edit" && event.preventDefault()}
    >
      <StopPropagation>
        <Checkbox
          defaultValue={value}
          fieldName={fieldName}
          // remount when defaultValue changes
          key={String(value)}
        />
      </StopPropagation>
    </Fieldset>
  );
};
