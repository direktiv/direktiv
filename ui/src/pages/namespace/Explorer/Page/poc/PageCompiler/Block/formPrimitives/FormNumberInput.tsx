import { Fieldset } from "./utils/FieldSet";
import { FormNumberInputType } from "../../../schema/blocks/form/numberInput";
import Input from "~/design/Input";
import { StopPropagation } from "~/components/StopPropagation";
import { encodeBlockKey } from "./utils";
import { useTranslation } from "react-i18next";
import { useVariableNumberResolver } from "../../primitives/Variable/utils/useVariableNumberResolver";

type FormNumberInputProps = {
  blockProps: FormNumberInputType;
};

export const FormNumberInput = ({ blockProps }: FormNumberInputProps) => {
  const { t } = useTranslation();
  const resolveVariableNumber = useVariableNumberResolver();
  const { id, label, description, defaultValue, optional, type } = blockProps;

  const fieldName = encodeBlockKey(type, id);
  let value: number;

  if (defaultValue.type === "variable") {
    const resolvedDefaultValue = resolveVariableNumber(defaultValue.value);
    if (!resolvedDefaultValue.success) {
      throw new Error(
        t(`direktivPage.error.templateString.${resolvedDefaultValue.error}`, {
          variable: defaultValue.value,
        })
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
      optional={optional}
    >
      <StopPropagation>
        <Input
          type="number"
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
