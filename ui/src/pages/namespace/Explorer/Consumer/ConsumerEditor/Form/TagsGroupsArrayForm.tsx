import { ArrayForm } from "~/components/Form/Array";
import { ConsumerFormSchemaType } from "../schema";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type TagsGroupsArrayFormProps = {
  field:
    | ControllerRenderProps<ConsumerFormSchemaType, "tags">
    | ControllerRenderProps<ConsumerFormSchemaType, "groups">;
};

export const TagsGroupsArrayForm = ({ field }: TagsGroupsArrayFormProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5 sm:grid-cols-2">
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem=""
        renderItem={({ value, setValue }) => (
          <Input
            placeholder={t(
              `pages.explorer.consumer.editor.form.${field.name}Placeholder`
            )}
            value={value}
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
