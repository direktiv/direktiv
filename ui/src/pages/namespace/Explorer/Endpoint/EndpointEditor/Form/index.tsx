import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
} from "react-hook-form";
import { EndpointFormSchemaType, EndpointSaveSchema } from "../schema";

import { AuthPluginForm } from "./plugins/Auth";
import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import { InboundPluginForm } from "./plugins/Inbound";
import Input from "~/design/Input";
import { MethodCheckbox } from "./methodCheckbox";
import { OutboundPluginForm } from "./plugins/Outbound";
import { Switch } from "~/design/Switch";
import { TargetPluginForm } from "./plugins/Target";
import { routeMethods } from "~/api/gateway/schema";
import { treatAsNumberOrUndefined } from "../../../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: EndpointFormSchemaType;
  onSave: (value: EndpointFormSchemaType) => void;
  children: (args: {
    formControls: UseFormReturn<EndpointFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<EndpointFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children, onSave }) => {
  const { t } = useTranslation();

  const formControls = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointSaveSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const values = formControls.watch();

  const { register, control } = formControls;

  return children({
    formControls,
    values,
    formMarkup: (
      <div className="flex flex-col gap-8">
        <div className="flex gap-3">
          <Fieldset
            label={t("pages.explorer.endpoint.editor.form.path")}
            htmlFor="path"
            className="grow"
          >
            <Input {...register("x-direktiv-config.path")} id="path" />
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
        <Fieldset label={t("pages.explorer.endpoint.editor.form.methods")}>
          <div className="grid grid-cols-3 gap-5">
            {Array.from(routeMethods).map((method) => (
              <Controller
                key={method}
                control={control}
                name={method}
                render={({ field }) => (
                  <MethodCheckbox method={method} field={field} />
                )}
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
        <TargetPluginForm form={formControls} onSave={onSave} />
        <InboundPluginForm form={formControls} onSave={onSave} />
        <OutboundPluginForm form={formControls} onSave={onSave} />
        <AuthPluginForm formControls={formControls} onSave={onSave} />
      </div>
    ),
  });
};
