import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";
import { VariableError } from "../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableBooleanResolver } from "../../primitives/Variable/utils/useVariableBooleanResolver";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-checkbox-${id}`;

  const resolveVariableBoolean = useVariableBooleanResolver();

  const isVariable = defaultValue.type === "variable";

  let value: boolean;
  if (isVariable) {
    const resolvedDefaultValue = resolveVariableBoolean(defaultValue.value);
    if (!resolvedDefaultValue.success) {
      return (
        <VariableError
          value={defaultValue.value}
          errorCode={resolvedDefaultValue.error}
        >
          {t(`direktivPage.error.templateString.${resolvedDefaultValue.error}`)}
        </VariableError>
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
    >
      <Checkbox defaultChecked={value} id={htmlID} />
    </Fieldset>
  );
};
