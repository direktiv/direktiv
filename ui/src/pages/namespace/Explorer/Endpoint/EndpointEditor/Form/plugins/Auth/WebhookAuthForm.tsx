import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  WebhookAuthFormSchema,
  WebhookAuthFormSchemaType,
} from "../../../schema/plugins/auth/webhookAuth";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { PluginWrapper } from "../components/PluginSelector";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<WebhookAuthFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  formId: string;
  type: WebhookAuthFormSchemaType["type"];
  onSubmit: (data: WebhookAuthFormSchemaType) => void;
};

export const WebhookAuthForm: FC<FormProps> = ({
  defaultConfig,
  formId,
  type,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<WebhookAuthFormSchemaType>({
    resolver: zodResolver(WebhookAuthFormSchema),
    defaultValues: {
      type,
      configuration: {
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
            "pages.explorer.endpoint.editor.form.plugins.auth.githubWebhookAuth.secret"
          )}
          htmlFor="secret"
        >
          <Input {...register("configuration.secret")} id="secret" />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
