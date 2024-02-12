import { AclFormSchemaType } from "../../../../schema/plugins/inbound/acl";
import { ArrayForm } from "~/components/Form/Array";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type AclArrayInputProps = {
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
      return "allow_groups";
    case "configuration.allow_tags":
      return "allow_tags";
    case "configuration.deny_groups":
      return "deny_groups";
    case "configuration.deny_tags":
      return "deny_tags";
  }
};

export const AclArrayInput = ({ field }: AclArrayInputProps) => {
  const { t } = useTranslation();
  return (
    <div className="grid gap-5 sm:grid-cols-2">
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem=""
        itemIsValid={(item) => item !== ""}
        renderItem={({ value, setValue, handleKeyDown }) => (
          <Input
            placeholder={t(
              `pages.explorer.endpoint.editor.form.plugins.inbound.acl.${fieldNameToLanguageKey(
                field.name
              )}`
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
