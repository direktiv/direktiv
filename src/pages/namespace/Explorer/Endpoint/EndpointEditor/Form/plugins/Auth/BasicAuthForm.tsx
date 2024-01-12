import {
  BasicAuthFormSchema,
  BasicAuthFormSchemaType,
} from "../../../schema/plugins/auth/basicAuth";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";

import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "~/pages/namespace/Explorer/components/Fieldset";
import { PluginWrapper } from "../components/Modal";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<BasicAuthFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
  add_groups_header: false,
  add_tags_header: false,
  add_username_header: false,
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: BasicAuthFormSchemaType) => void;
};

export const BasicAuthForm: FC<FormProps> = ({
  defaultConfig,
  formId,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<BasicAuthFormSchemaType>({
    resolver: zodResolver(BasicAuthFormSchema),
    defaultValues: {
      type: "basic-auth",
      configuration: {
        ...predefinedConfig,
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm} id={formId}>
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.basciAuth.addUsernameHeader"
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
            "pages.explorer.endpoint.editor.form.plugins.auth.basciAuth.addTagsHeader"
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
            "pages.explorer.endpoint.editor.form.plugins.auth.basciAuth.addGroupsHeader"
          )}
          horizontal
          htmlFor="add-groups-header"
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
      </PluginWrapper>
    </form>
  );
};
