import { Fieldset } from "./utils/FieldSet";
import { FormTextareaType } from "../../../schema/blocks/form/textarea";
import { Textarea } from "~/design/TextArea";
import { VariableError } from "../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "../../primitives/Variable/utils/useVariableStringResolver";

type FormTextareaProps = {
  blockProps: FormTextareaType;
};

export const FormTextarea = ({ blockProps }: FormTextareaProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-textarea-${id}`;

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
      <Textarea defaultValue={resolvedDefaultValue.data} id={htmlID} />
    </Fieldset>
  );
};
