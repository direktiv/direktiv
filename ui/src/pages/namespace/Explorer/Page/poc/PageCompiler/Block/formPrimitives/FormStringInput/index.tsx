import { DateInput } from "./DateInput";
import { Fieldset } from "../utils/FieldSet";
import { FormStringInputType } from "../../../../schema/blocks/form/stringInput";
import { TextInput } from "./TextInput";
import { VariableError } from "../../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "../../../primitives/Variable/utils/useVariableStringResolver";

type FormInputProps = {
  blockProps: FormStringInputType;
};

export const FormStringInput = ({ blockProps }: FormInputProps) => {
  const { t } = useTranslation();
  const { id, label, description, variant, defaultValue, optional } =
    blockProps;
  const htmlID = `form-input-${id}`;

  const resolveVariableString = useVariableStringResolver();
  const resolvedDefaultValue = resolveVariableString(defaultValue);

  if (!resolvedDefaultValue.success) {
    return (
      <VariableError
        value={defaultValue}
        errorCode={resolvedDefaultValue.error}
      >
        {t(`direktivPage.error.templateString.${resolvedDefaultValue.error}`)}
      </VariableError>
    );
  }

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
          defaultValue={resolvedDefaultValue.data}
          // remount when defaultValue changes
          key={defaultValue}
        />
      ) : (
        <TextInput
          id={htmlID}
          defaultValue={resolvedDefaultValue.data}
          variant={variant}
        />
      )}
    </Fieldset>
  );
};
