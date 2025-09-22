import { AclFormSchemaType } from "../../../../schema/plugins/inbound/acl";
import { ArrayForm } from "~/components/Form/Array";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type AclArrayFormProps = {
  field:
    | ControllerRenderProps<AclFormSchemaType, "configuration.allow_groups">
    | ControllerRenderProps<AclFormSchemaType, "configuration.allow_tags">
    | ControllerRenderProps<AclFormSchemaType, "configuration.deny_groups">
    | ControllerRenderProps<AclFormSchemaType, "configuration.deny_tags">;
};

const fieldNameToLanguageKey = (
  field:
    | "configuration.allow_groups"
    | "configuration.allow_tags"
    | "configuration.deny_groups"
    | "configuration.deny_tags"
) => {
  switch (field) {
    case "configuration.allow_groups":
    case "configuration.deny_groups":
      return "groupPlaceholder";
    case "configuration.allow_tags":
    case "configuration.deny_tags":
      return "tagPlaceholder";
  }
};

export const AclArrayForm = ({ field }: AclArrayFormProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5 sm:grid-cols-2">
      <ArrayForm
        value={field.value || []}
        onChange={field.onChange}
        emptyItem=""
        renderItem={({ value, setValue }) => (
          <Input
            placeholder={t(
              `pages.explorer.endpoint.editor.form.plugins.inbound.acl.${fieldNameToLanguageKey(
                field.name
              )}`
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
