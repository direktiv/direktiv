import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";
import { StopPropagation } from "~/components/StopPropagation";
import { usePageStateContext } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";
import { useVariableBooleanResolver } from "../../primitives/Variable/utils/useVariableBooleanResolver";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const { t } = useTranslation();
  const resolveVariableBoolean = useVariableBooleanResolver();
  const { id, label, description, defaultValue, optional } = blockProps;
  const { mode } = usePageStateContext();

  const htmlID = `form-checkbox-${id}`;
  let value: boolean;

  if (defaultValue.type === "variable") {
    const resolvedDefaultValue = resolveVariableBoolean(defaultValue.value);
    if (!resolvedDefaultValue.success) {
      throw new Error(
        t(`direktivPage.error.templateString.${resolvedDefaultValue.error}`)
      );
    }
    value = resolvedDefaultValue.data;
  } else {
    value = defaultValue.value;
  }

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      horizontal
      optional={optional}
      onClickLabel={(event) => mode === "edit" && event.preventDefault()}
    >
      <StopPropagation asChild>
        <Checkbox
          defaultChecked={value}
          id={htmlID}
          // remount when defaultValue changes
          key={String(value)}
        />
      </StopPropagation>
    </Fieldset>
  );
};
