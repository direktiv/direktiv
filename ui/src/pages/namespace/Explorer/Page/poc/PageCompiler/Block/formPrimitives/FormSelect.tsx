import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";
import { useTranslation } from "react-i18next";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, values, optional } = blockProps;
  const htmlID = `id-${id}`;

  const value = values.some((v) => v === defaultValue)
    ? defaultValue
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
