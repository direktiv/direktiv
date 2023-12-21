import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  KeyAuthFormSchema,
  KeyAuthFormSchemaType,
} from "../../../schema/plugins/auth/keyAuth";

import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "../../components/FormHelper";
import Input from "~/design/Input";
import { PluginWrapper } from "../components/Modal";
import { treatEmptyStringAsUndefined } from "../utils";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<KeyAuthFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  add_groups_header: true,
  add_tags_header: true,
  add_username_header: true,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: KeyAuthFormSchemaType) => void;
};

export const KeyAuthForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    setValue,
    getValues,
    register,
    formState: { errors },
  } = useForm<KeyAuthFormSchemaType>({
    resolver: zodResolver(KeyAuthFormSchema),
    defaultValues: {
      type: "key-auth",
      configuration: {
        ...predfinedConfig,
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm}>
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.keyAuth.addUsernameHeader"
          )}
          htmlFor="add-username-header"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.add_username_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_username_header", value);
              }
            }}
            id="add-username-header"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.keyAuth.addTagsHeader"
          )}
          htmlFor="add-tags-header"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.add_tags_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_tags_header", value);
              }
            }}
            id="add-tags-header"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.keyAuth.addGroupsHeader"
          )}
          htmlFor="add-groups-header"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.add_groups_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_groups_header", value);
              }
            }}
            id="add-groups-header"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.keyAuth.keyName"
          )}
          htmlFor="key-name"
        >
          <Input
            {...register("configuration.key_name", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.auth.keyAuth.keyNamePlaceholder"
            )}
            id="key-name"
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
