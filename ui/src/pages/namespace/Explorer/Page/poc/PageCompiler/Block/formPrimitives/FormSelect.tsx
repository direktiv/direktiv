import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";
import { StopPropagation } from "~/components/StopPropagation";
import { encodeBlockKey } from "./utils";
import { getStringFromJsonPath } from "../../primitives/Variable/utils";
import { useStringInterpolation } from "../../primitives/Variable/utils/useStringInterpolation";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../../primitives/Variable/utils/useVariableArrayResolver";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const interpolateString = useStringInterpolation();
  const variableResolver = useVariableArrayResolver();
  const { id, label, description, defaultValue, values, optional, type } =
    blockProps;

  const resolvedDefaultValue = interpolateString(defaultValue);
  const fieldName = encodeBlockKey(type, id, optional);

  let resolvedValues: { value: string; label: string }[];

  if (values.type === "variable-select-options") {
    const result = variableResolver(values.arrayPath);
    if (result.success) {
      resolvedValues = result.data.map((object) => {
        const value = getStringFromJsonPath(object, values.valuePath);
        const label = getStringFromJsonPath(object, values.labelPath);
        return { value, label };
      });
    } else {
      throw new Error(
        t(`direktivPage.error.templateString.${result.error}`, {
          variable: values.arrayPath,
        })
      );
    }
  } else {
    resolvedValues = values.value.map((value) => ({ label: value, value }));
  }

  const value = resolvedValues.find((v) => v.label === resolvedDefaultValue);

  return (
    <Fieldset
      id={id}
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <Select defaultValue={value?.label} key={value?.label} name={fieldName}>
        <StopPropagation>
          <SelectTrigger variant="outline" id={fieldName} value={value?.label}>
            <SelectValue
              placeholder={t("direktivPage.page.blocks.form.selectPlaceholder")}
            />
          </SelectTrigger>
        </StopPropagation>
        <SelectContent>
          {resolvedValues.map(({ value, label }) => (
            <SelectItem key={value} value={value}>
              {label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </Fieldset>
  );
};
