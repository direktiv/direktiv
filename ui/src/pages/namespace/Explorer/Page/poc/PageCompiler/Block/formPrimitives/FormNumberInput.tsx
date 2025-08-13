import { Fieldset } from "./utils/FieldSet";
import { FormNumberInputType } from "../../../schema/blocks/form/numberInput";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { useVariableNumberResolver } from "../../primitives/Variable/utils/useVariableNumberResolver";

type FormNumberInputProps = {
  blockProps: FormNumberInputType;
};

export const FormNumberInput = ({ blockProps }: FormNumberInputProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-input-${id}`;

  const resolveVariableNumber = useVariableNumberResolver();

  let value: number;

  if (defaultValue.type === "variable") {
    const resolvedDefaultValue = resolveVariableNumber(defaultValue.value);
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
      optional={optional}
    >
      <Input
        type="number"
        defaultValue={value}
        id={htmlID}
        name={id}
        // remount when defaultValue changes
        key={value}
      />
    </Fieldset>
  );
};
