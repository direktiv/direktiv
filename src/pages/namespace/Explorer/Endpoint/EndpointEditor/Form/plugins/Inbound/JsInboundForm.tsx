import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  JsInboundFormSchema,
  JsInboundFormSchemaType,
} from "../../../schema/plugins/inbound/jsInbound";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Fieldset } from "~/pages/namespace/Explorer/components/Fieldset";
import { PluginWrapper } from "../components/Modal";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<JsInboundFormSchemaType["configuration"]>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: JsInboundFormSchemaType) => void;
};

export const JsInboundForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<JsInboundFormSchemaType>({
    resolver: zodResolver(JsInboundFormSchema),
    defaultValues: {
      type: "js-inbound",
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
            "pages.explorer.endpoint.editor.form.plugins.inbound.jsInbound.script"
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
