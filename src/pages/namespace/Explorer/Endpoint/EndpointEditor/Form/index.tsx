import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";
import { EndpointFormSchema, EndpointFormSchemaType } from "../schema";

import { AuthPluginForm } from "./plugins/Auth";
import Badge from "~/design/Badge";
import { Checkbox } from "~/design/Checkbox";
import { FC } from "react";
import { Fieldset } from "../../../components/Fieldset";
import { InboundPluginForm } from "./plugins/Inbound";
import Input from "~/design/Input";
import { OutboundPluginForm } from "./plugins/Outbound";
import { Switch } from "~/design/Switch";
import { TargetPluginForm } from "./plugins/Target";
import { routeMethods } from "~/api/gateway/schema";
import { treatAsNumberOrUndefined } from "../../../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: EndpointFormSchemaType;
  children: (args: {
    formControls: UseFormReturn<EndpointFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<EndpointFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const formControls = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const values = useWatch({
    control: formControls.control,
  });

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
            <Input {...register("path")} id="path" />
          </Fieldset>
          <Fieldset
            label={t("pages.explorer.endpoint.editor.form.timeout")}
            htmlFor="timeout"
            className="w-32"
          >
            <Input
              {...register("timeout", {
                setValueAs: treatAsNumberOrUndefined,
              })}
              type="number"
              id="timeout"
            />
          </Fieldset>
        </div>
        <Fieldset label={t("pages.explorer.endpoint.editor.form.methods")}>
          <Controller
            control={control}
            name="methods"
            render={({ field }) => (
              <div className="grid grid-cols-3 gap-5">
                {routeMethods.map((method) => {
                  const isChecked = field.value?.includes(method);
                  return (
                    <label
                      key={method}
                      className="flex items-center gap-2 text-sm"
                      htmlFor={method}
                    >
                      <Checkbox
                        id={method}
                        value={method}
                        checked={isChecked}
                        onCheckedChange={(checked) => {
                          if (checked === true) {
                            field.onChange([...(field.value ?? []), method]);
                          }
                          if (checked === false && field.value) {
                            field.onChange(
                              field.value.filter((v) => v !== method)
                            );
                          }
                        }}
                      />
                      <Badge variant={isChecked ? undefined : "secondary"}>
                        {method}
                      </Badge>
                    </label>
                  );
                })}
              </div>
            )}
          />
        </Fieldset>
        <Fieldset
          label={t("pages.explorer.endpoint.editor.form.allowAnonymous")}
          htmlFor="allow_anonymous"
          horizontal
        >
          <Controller
            control={control}
            name="allow_anonymous"
            render={({ field }) => (
              <Switch
                defaultChecked={field.value ?? false}
                onCheckedChange={(value) => {
                  field.onChange(value);
                }}
                id="allow_anonymous"
              />
            )}
          />
        </Fieldset>
        <TargetPluginForm form={formControls} />
        <InboundPluginForm form={formControls} />
        <OutboundPluginForm form={formControls} />
        <AuthPluginForm formControls={formControls} />
      </div>
    ),
  });
};
