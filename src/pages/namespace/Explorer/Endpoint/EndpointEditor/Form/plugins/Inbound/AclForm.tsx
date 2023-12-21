import {
  AclFormSchema,
  AclFormSchemaType,
} from "../../../schema/plugins/inbound/acl";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import { ModalFooter, PluginWrapper } from "../components/Modal";

import { Fieldset } from "../../components/FormHelper";
import { useForm } from "react-hook-form";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<AclFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: AclFormSchemaType) => void;
};

export const AclForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
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
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  const theme = useTheme();

  return (
    <form onSubmit={submitForm}>
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
          ...
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.deny_groups"
          )}
        >
          ...
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.allow_tags"
          )}
        >
          ...
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.acl.deny_tags"
          )}
        >
          ...
        </Fieldset>
      </PluginWrapper>
      <ModalFooter />
    </form>
  );
};
