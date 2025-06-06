import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";
import { EndpointFormSchema, EndpointFormSchemaType } from "../schema";

import { AuthPluginForm } from "./plugins/Auth";
import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import { InboundPluginForm } from "./plugins/Inbound";
import Input from "~/design/Input";
import { MethodCheckbox } from "./MethodCheckbox";
import { OpenAPIDocsForm } from "./openAPIDocs";
import { OutboundPluginForm } from "./plugins/Outbound";
import { Switch } from "~/design/Switch";
import { TargetPluginForm } from "./plugins/Target";
import { forceLeadingSlash } from "~/api/files/utils";
import { routeMethods } from "~/api/gateway/schema";
import { treatAsNumberOrUndefined } from "../../../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: DeepPartialSkipArrayKey<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
  children: (args: {
    form: UseFormReturn<EndpointFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<EndpointFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children, onSave }) => {
  const { t } = useTranslation();

  const form = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const values = useWatch({ control: form.control });

  const { register, control } = form;

  return children({
    form,
    values,
    formMarkup: (
      <div className="flex flex-col gap-6">
        <div className="flex gap-3">
          <Fieldset
            label={t("pages.explorer.endpoint.editor.form.path")}
            htmlFor="path"
            className="grow"
          >
            <Input
              {...register("x-direktiv-config.path")}
              id="path"
              onChange={(event) =>
                form.setValue(
                  "x-direktiv-config.path",
                  forceLeadingSlash(event.target.value)
                )
              }
            />
          </Fieldset>
          <Fieldset
            label={t("pages.explorer.endpoint.editor.form.timeout")}
            htmlFor="timeout"
            className="w-32"
          >
            <Input
              {...register("x-direktiv-config.timeout", {
                setValueAs: treatAsNumberOrUndefined,
              })}
              type="number"
              id="timeout"
            />
          </Fieldset>
        </div>
        <Fieldset
          label={t("pages.explorer.endpoint.editor.form.methods.label")}
        >
          <div className="grid grid-cols-3 gap-5">
            {Array.from(routeMethods).map((method) => (
              <Controller
                key={method}
                control={control}
                name={method}
                render={({ field }) => {
                  const isChecked = !!values[method];
                  return (
                    <div className="flex items-center gap-2">
                      <MethodCheckbox
                        method={method}
                        field={field}
                        isChecked={isChecked}
                        form={form}
                      />
                    </div>
                  );
                }}
              />
            ))}
          </div>
        </Fieldset>
        <Fieldset
          label={t("pages.explorer.endpoint.editor.form.allowAnonymous")}
          htmlFor="x-direktiv-config.allow_anonymous"
          horizontal
        >
          <Controller
            control={control}
            name="x-direktiv-config.allow_anonymous"
            render={({ field }) => (
              <Switch
                checked={field.value ?? false}
                onCheckedChange={(value) => {
                  field.onChange(value);
                }}
                id="x-direktiv-config.allow_anonymous"
              />
            )}
          />
        </Fieldset>
        <TargetPluginForm form={form} onSave={onSave} />
        <InboundPluginForm form={form} onSave={onSave} />
        <OutboundPluginForm form={form} onSave={onSave} />
        <AuthPluginForm form={form} onSave={onSave} />
        <OpenAPIDocsForm form={form} onSave={onSave} />
      </div>
    ),
  });
};
