import {
  AclFormSchema,
  AclFormSchemaType,
} from "../../../schema/plugins/inbound/acl";
import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";

import { Fieldset } from "../../components/FormHelper";
import { GatewayArray } from "~/design/GatewayForms";
import { PluginWrapper } from "../components/Modal";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<AclFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  allow_groups: [],
  deny_groups: [],
  allow_tags: [],
  deny_tags: [],
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: AclFormSchemaType) => void;
};

export const AclForm: FC<FormProps> = ({ defaultConfig, onSubmit, formId }) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<AclFormSchemaType>({
    resolver: zodResolver(AclFormSchema),
    defaultValues: {
      type: "acl",
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
    <form onSubmit={submitForm} id={formId}>
      {errors?.configuration && (
        <FormErrors
          errors={errors?.configuration as errorsType}
          className="mb-5"
        />
      )}
      <PluginWrapper>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.allow_groups"
          )}
        >
          <Controller
            control={control}
            name="configuration.allow_groups"
            render={({ field }) => (
              <GatewayArray
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.inbound.acl.groupPlaceholder"
                )}
                externalArray={field.value ?? []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.deny_groups"
          )}
        >
          <Controller
            control={control}
            name="configuration.deny_groups"
            render={({ field }) => (
              <GatewayArray
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.inbound.acl.groupPlaceholder"
                )}
                externalArray={field.value ?? []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.allow_tags"
          )}
        >
          <Controller
            control={control}
            name="configuration.allow_tags"
            render={({ field }) => (
              <GatewayArray
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.inbound.acl.tagPlaceholder"
                )}
                externalArray={field.value ?? []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.deny_tags"
          )}
        >
          <Controller
            control={control}
            name="configuration.deny_tags"
            render={({ field }) => (
              <GatewayArray
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.inbound.acl.tagPlaceholder"
                )}
                externalArray={field.value ?? []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
