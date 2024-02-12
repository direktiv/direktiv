import {
  AclFormSchema,
  AclFormSchemaType,
} from "../../../../schema/plugins/inbound/acl";
import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { AclArrayForm } from "./AclArrayForm";
import { Fieldset } from "~/components/Form/Fieldset";
import { PluginWrapper } from "../../components/PluginSelector";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<AclFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
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
            render={({ field }) => <AclArrayForm field={field} />}
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
            render={({ field }) => <AclArrayForm field={field} />}
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
            render={({ field }) => <AclArrayForm field={field} />}
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
            render={({ field }) => <AclArrayForm field={field} />}
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
