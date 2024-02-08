import { ArrayInput } from "~/components/Form/ArrayInput";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { ServiceFormSchemaType } from "../schema";
import { useTranslation } from "react-i18next";

type EnvsArrayInputProps = {
  field: ControllerRenderProps<ServiceFormSchemaType, "envs">;
};

export const EnvsArrayInput = ({ field }: EnvsArrayInputProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5" data-testid="env-item-form">
      <ArrayInput
        defaultValue={field.value || []}
        onChange={(changedValue) => {
          field.onChange(changedValue);
        }}
        emptyItem={{ name: "", value: "" }}
        itemIsValid={(item) => !!(item?.name && item?.value)}
        renderItem={({ value: objectValue, setValue, handleKeyDown }) => (
          <>
            {Object.entries(objectValue).map(([key, value]) => {
              const typedKey = key as keyof typeof objectValue;
              return (
                <Input
                  key={key}
                  data-testid={`env-${typedKey}`}
                  placeholder={t(
                    `pages.explorer.service.editor.form.envs.${typedKey}Placeholder`
                  )}
                  value={value}
                  onKeyDown={handleKeyDown}
                  onChange={(e) => {
                    // TODO: better naming
                    const newObject = { [key]: e.target.value };
                    const newObjectValue = {
                      ...objectValue,
                      ...newObject,
                    };
                    setValue(newObjectValue);
                  }}
                />
              );
            })}
          </>
        )}
      />
    </div>
  );
};
