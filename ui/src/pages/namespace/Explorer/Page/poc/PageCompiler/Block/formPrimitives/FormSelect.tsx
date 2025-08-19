import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";
import { serializeFieldName } from "./utils";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";
import { useTranslation } from "react-i18next";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const templateStringResolver = useTemplateStringResolver();
  const { id, label, description, defaultValue, values, optional, type } =
    blockProps;

  const resolvedDefaultValue = templateStringResolver(defaultValue);
  const fieldName = serializeFieldName(type, id);
  const value = values.some((v) => v === resolvedDefaultValue)
    ? resolvedDefaultValue
    : undefined;

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <Select
        defaultValue={value}
        // remount when defaultValue changes
        key={value}
        name={fieldName}
      >
        <SelectTrigger variant="outline" id={fieldName} value={value}>
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
