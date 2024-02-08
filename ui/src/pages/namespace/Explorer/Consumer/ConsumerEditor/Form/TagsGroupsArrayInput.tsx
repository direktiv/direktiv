import { ArrayInput } from "~/components/Form/ArrayInput";
import { ConsumerFormSchemaType } from "../schema";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type TagsGroupsArrayInputProps = {
  field:
    | ControllerRenderProps<ConsumerFormSchemaType, "tags">
    | ControllerRenderProps<ConsumerFormSchemaType, "groups">;
};

export const TagsGroupsArrayInput = ({ field }: TagsGroupsArrayInputProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5 sm:grid-cols-2">
      <ArrayInput
        defaultValue={field.value || []}
        onChange={(changedValue) => {
          field.onChange(changedValue);
        }}
        emptyItem=""
        itemIsValid={(item) => item !== ""}
        renderItem={({ value, setValue, handleKeyDown }) => (
          <Input
            placeholder={t(
              `pages.explorer.consumer.editor.form.${field.name}Placeholder`
            )}
            value={value}
            onKeyDown={handleKeyDown}
            onChange={(e) => {
              const newValue = e.target.value;
              setValue(newValue);
            }}
          />
        )}
      />
    </div>
  );
};
