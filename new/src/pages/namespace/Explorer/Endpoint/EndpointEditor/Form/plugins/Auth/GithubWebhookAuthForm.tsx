import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  GithubWebhookAuthFormSchema,
  GithubWebhookAuthFormSchemaType,
} from "../../../schema/plugins/auth/githubWebhookAuth";

import { Fieldset } from "~/pages/namespace/Explorer/components/Fieldset";
import Input from "~/design/Input";
import { PluginWrapper } from "../components/Modal";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<GithubWebhookAuthFormSchemaType["configuration"]>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: GithubWebhookAuthFormSchemaType) => void;
};

export const GithubWebhookAuthForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<GithubWebhookAuthFormSchemaType>({
    resolver: zodResolver(GithubWebhookAuthFormSchema),
    defaultValues: {
      type: "github-webhook-auth",
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
