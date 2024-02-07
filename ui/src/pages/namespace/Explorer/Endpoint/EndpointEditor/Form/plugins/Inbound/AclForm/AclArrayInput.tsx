import { AclFormSchemaType } from "../../../../schema/plugins/inbound/acl";
import { ArrayInput } from "~/components/Form/ArrayInput";
import { ControllerRenderProps } from "react-hook-form";
import Input from "~/design/Input";

type AclArrayInputProps = {
  field:
    | ControllerRenderProps<AclFormSchemaType, "configuration.allow_groups">
    | ControllerRenderProps<AclFormSchemaType, "configuration.allow_tags">
    | ControllerRenderProps<AclFormSchemaType, "configuration.deny_groups">
    | ControllerRenderProps<AclFormSchemaType, "configuration.deny_tags">;
  placeholder: string;
};

export const AclArrayInput = ({ field, placeholder }: AclArrayInputProps) => (
  <div className="grid gap-5 sm:grid-cols-2">
    <ArrayInput
      defaultValue={field.value || []}
      onChange={(changedValue) => {
        field.onChange(changedValue);
      }}
      emptyItem=""
      itemIsValid={(item) => item !== ""}
      renderItem={({ value, setValue, onChange, handleKeyDown }) => (
        <Input
          placeholder={placeholder}
          value={value}
          onKeyDown={handleKeyDown}
          onChange={(e) => {
            const newValue = e.target.value;
            setValue(newValue);
            onChange(newValue);
          }}
        />
      )}
    />
  </div>
);
