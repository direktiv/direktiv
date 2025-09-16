import { ArrayForm } from "~/components/Form/Array";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { useTranslation } from "react-i18next";

type KeyValueInputProps = {
  label: string;
  field: {
    value: KeyValueType[] | undefined;
    onChange: (value: KeyValueType[]) => void;
  };
};

export const KeyValueInput = ({ field, label }: KeyValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset label={label}>
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem={{ key: "", value: "" }}
        renderItem={({ value: itemValue, setValue }) => (
          <>
            <Input
              placeholder={t(
                "direktivPage.blockEditor.blockForms.keyValue.key"
              )}
              value={itemValue.key}
              onChange={(e) => {
                setValue({
                  ...itemValue,
                  key: e.target.value,
                });
              }}
            />
            <Input
              placeholder={t(
                "direktivPage.blockEditor.blockForms.keyValue.value"
              )}
              value={itemValue.value}
              onChange={(e) => {
                setValue({
                  ...itemValue,
                  value: e.target.value,
                });
              }}
            />
          </>
        )}
      />
    </Fieldset>
  );
};
