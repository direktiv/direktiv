import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "./utils/FieldSet";
import { FormCheckboxType } from "../../../schema/blocks/form/checkbox";
import { serializeFieldName } from "./utils";
import { useTranslation } from "react-i18next";
import { useVariableBooleanResolver } from "../../primitives/Variable/utils/useVariableBooleanResolver";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => {
  const { t } = useTranslation();
  const resolveVariableBoolean = useVariableBooleanResolver();
  const { id, label, description, defaultValue, optional, type } = blockProps;

  const fieldName = serializeFieldName(type, id);
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
      htmlFor={fieldName}
      horizontal
      optional={optional}
    >
      <Checkbox
        defaultChecked={value}
        id={fieldName}
        name={fieldName}
        // remount when defaultValue changes
        key={String(value)}
      />
    </Fieldset>
  );
};
