import { ArrayForm } from "~/components/Form/Array";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { SmartInput } from "../SmartInput";
import { useTranslation } from "react-i18next";

type KeyValueInputProps = {
  label: string;
  field: {
    value: KeyValueType[] | undefined;
    onChange: (value: KeyValueType[]) => void;
  };
  smart?: boolean;
};

export const KeyValueInput = ({
  field,
  label,
  smart = false,
}: KeyValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset label={label}>
      <ArrayForm
        value={field.value || []}
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
            {smart ? (
              <SmartInput
                placeholder={t(
                  "direktivPage.blockEditor.blockForms.keyValue.value"
                )}
                value={itemValue.value}
                onUpdate={(content) => {
                  setValue({
                    ...itemValue,
                    value: content,
                  });
                }}
              />
            ) : (
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
            )}
          </>
        )}
      />
    </Fieldset>
  );
};
