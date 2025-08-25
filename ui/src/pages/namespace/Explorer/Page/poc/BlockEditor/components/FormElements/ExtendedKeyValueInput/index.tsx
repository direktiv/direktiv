import {
  ExtendedKeyValue,
  ExtendedKeyValueType,
} from "../../../../schema/primitives/extendedKeyValue";

import { ArrayForm } from "~/components/Form/Array";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { ValueInput } from "./ValueInput";
import { useTranslation } from "react-i18next";

type ExtendedKeyValueInputProps = {
  label: string;
  field: {
    value: ExtendedKeyValueType[] | undefined;
    onChange: (value: ExtendedKeyValueType[]) => void;
  };
};

export const ExtendedKeyValueInput = ({
  field,
  label,
}: ExtendedKeyValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset label={label}>
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem={{
          key: "",
          value: {
            type: "string",
            value: "",
          },
        }}
        itemIsValid={(item) => ExtendedKeyValue.safeParse(item).success}
        renderItem={({ value: itemValue, setValue, handleKeyDown }) => (
          <div className="flex flex-col gap-2">
            <Input
              placeholder={t(
                "direktivPage.blockEditor.blockForms.keyValue.key"
              )}
              value={itemValue.key}
              onKeyDown={handleKeyDown}
              onChange={(e) => {
                setValue({
                  ...itemValue,
                  key: e.target.value,
                });
              }}
            />
            <ValueInput
              value={itemValue.value}
              onChange={(newValue) => {
                setValue({
                  ...itemValue,
                  value: newValue,
                });
              }}
              onKeyDown={handleKeyDown}
            />
          </div>
        )}
      />
    </Fieldset>
  );
};
