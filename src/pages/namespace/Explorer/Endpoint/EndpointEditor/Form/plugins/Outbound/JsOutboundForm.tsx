import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  JsOutboundFormSchema,
  JsOutboundFormSchemaType,
} from "../../../schema/plugins/outbound/jsOutbound";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Fieldset } from "../../components/FormHelper";
import { PluginWrapper } from "../components/Modal";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<JsOutboundFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: JsOutboundFormSchemaType) => void;
};

export const JsOutboundForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<JsOutboundFormSchemaType>({
    resolver: zodResolver(JsOutboundFormSchema),
    defaultValues: {
      type: "js-outbound",
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
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.outbound.jsOutbound.script"
          )}
        >
          <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
            <Controller
              control={control}
              name="configuration.script"
              render={({ field }) => (
                <Editor
                  theme={theme ?? undefined}
                  language="javascript"
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
