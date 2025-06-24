import { KeyValue, KeyValueType } from "../../../schema/primitives/keyValue";

import { ArrayForm } from "~/components/Form/Array";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
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
      <ArrayForm<KeyValueType>
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem={{ key: "", value: "" }}
        itemIsValid={(item) => KeyValue.safeParse(item).success}
        renderItem={({ value: objectValue, setValue, handleKeyDown }) => (
          <>
            <Input
              placeholder={t(
                "direktivPage.blockEditor.blockForms.keyValue.key"
              )}
              value={objectValue.key}
              onKeyDown={handleKeyDown}
              onChange={(e) => {
                setValue({
                  ...objectValue,
                  key: e.target.value,
                });
              }}
            />
            <Input
              placeholder={t(
                "direktivPage.blockEditor.blockForms.keyValue.value"
              )}
              value={objectValue.value}
              onKeyDown={handleKeyDown}
              onChange={(e) => {
                setValue({
                  ...objectValue,
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
