import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  InstantResponseFormSchema,
  InstantResponseFormSchemaType,
} from "../../../schema/plugins/target/instantResponse";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { PluginWrapper } from "../components/Modal";
import { treatEmptyStringAsUndefined } from "~/pages/namespace/Explorer/utils";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<InstantResponseFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
  status_code: 200,
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: InstantResponseFormSchemaType) => void;
};

export const InstantResponseForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<InstantResponseFormSchemaType>({
    resolver: zodResolver(InstantResponseFormSchema),
    defaultValues: {
      type: "instant-response",
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

  const theme = useTheme();

  return (
    <form onSubmit={submitForm} id={formId}>
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors errors={errors?.configuration as errorsType} />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.instantResponse.statusCode"
          )}
          htmlFor="status-code"
        >
          <Input
            {...register("configuration.status_code", {
              valueAsNumber: true,
            })}
            id="status-code"
            type="number"
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.instantResponse.statusCodePlaceholder"
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.instantResponse.contentType"
          )}
          htmlFor="content-type"
        >
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            id="content-type"
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.instantResponse.contentTypePlaceholder"
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.instantResponse.statusMessage"
          )}
        >
          <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
            <Controller
              control={control}
              name="configuration.status_message"
              render={({ field }) => (
                <Editor
                  theme={theme ?? undefined}
                  language="plaintext"
                  value={field.value}
                  onChange={field.onChange}
                />
              )}
            />
          </Card>
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
