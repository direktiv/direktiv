import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";
import { useTemplateStringResolver } from "../../primitives/Variable/utils/useTemplateStringResolver";
import { useTranslation } from "react-i18next";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, values, optional } = blockProps;
  const htmlID = `form-select-${id}`;

  const templateStringResolver = useTemplateStringResolver();
  const resolvedDefaultValue = templateStringResolver(defaultValue);

  const value = values.some((v) => v === resolvedDefaultValue)
    ? resolvedDefaultValue
    : undefined;

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Select
        defaultValue={value}
        // remount when defaultValue changes
        key={value}
      >
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
