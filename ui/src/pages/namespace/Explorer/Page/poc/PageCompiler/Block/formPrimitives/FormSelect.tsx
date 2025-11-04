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
import { getStringValueFromJsonPath } from "../../primitives/Variable/utils";
import { useStringInterpolation } from "../../primitives/Variable/utils/useStringInterpolation";
import { useTranslation } from "react-i18next";
import { useUnwrapOrThrow } from "../../primitives/Variable/utils/useUnwrapOrThrow";
import { useVariableArrayResolver } from "../../primitives/Variable/utils/useVariableArrayResolver";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { t } = useTranslation();
  const unwrapOrThrow = useUnwrapOrThrow();
  const interpolateString = useStringInterpolation();
  const variableResolver = useVariableArrayResolver();
  const { id, label, description, defaultValue, values, optional, type } =
    blockProps;

  const resolvedDefaultValue = interpolateString(defaultValue);
  const fieldName = encodeBlockKey(type, id, optional);

  let resolvedValues: { value: string; label: string }[];

  if (values.type === "variable-select-options") {
    const arrayResult = variableResolver(values.arrayPath);
    const resolvedArray = unwrapOrThrow(arrayResult, values.arrayPath);

    resolvedValues = resolvedArray.map((object) => {
      const labelResult = getStringValueFromJsonPath(object, values.labelPath);
      const valueResult = getStringValueFromJsonPath(object, values.valuePath);

      return {
        label: unwrapOrThrow(labelResult, values.labelPath),
        value: unwrapOrThrow(valueResult, values.valuePath),
      };
    });
  } else {
    resolvedValues = values.value.map((value) => ({ label: value, value }));
  }

  const value = resolvedValues.find((v) => v.value === resolvedDefaultValue);

  return (
    <Fieldset
      id={id}
      label={label}
      description={description}
      htmlFor={fieldName}
      optional={optional}
    >
      <Select defaultValue={value?.value} key={value?.value} name={fieldName}>
        <StopPropagation>
          <SelectTrigger variant="outline" id={fieldName} value={value?.value}>
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
