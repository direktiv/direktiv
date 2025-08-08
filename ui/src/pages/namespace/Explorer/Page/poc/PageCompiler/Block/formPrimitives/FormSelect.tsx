import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";
import { VariableError } from "../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableStringResolver } from "../../primitives/Variable/utils/useVariableStringResolver";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, values, optional } = blockProps;
  const htmlID = `form-select-${id}`;

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

  const value = values.some((v) => v === resolvedDefaultValue.data)
    ? resolvedDefaultValue.data
    : undefined;

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Select defaultValue={value}>
        <SelectTrigger variant="outline" id={htmlID}>
          <SelectValue
            placeholder={t("direktivPage.page.blocks.form.selectPlaceholder")}
          />
        </SelectTrigger>
        <SelectContent>
          {values.map((value) => (
            <SelectItem key={value} value={value}>
              {value}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </Fieldset>
  );
};
