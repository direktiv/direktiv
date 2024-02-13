import { ArrayForm } from "~/components/Form/Array";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { ServiceFormSchemaType } from "../schema";
import { useTranslation } from "react-i18next";

type EnvsArrayFormProps = {
  field: ControllerRenderProps<ServiceFormSchemaType, "envs">;
};

export const EnvsArrayForm = ({ field }: EnvsArrayFormProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5" data-testid="env-item-form">
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem={{ name: "", value: "" }}
        itemIsValid={(item) =>
          !!item && Object.values(item).every((v) => v !== "")
        }
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
                    const newObject = {
                      ...objectValue,
                      [key]: e.target.value,
                    };
                    setValue(newObject);
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
