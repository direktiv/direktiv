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
import { InboundPluginForm } from "./plugins/Inbound";
import Input from "~/design/Input";
import { OutboundPluginForm } from "./plugins/Outbound";
import { Switch } from "~/design/Switch";
import { TargetPluginForm } from "./plugins/Target";
import { routeMethods } from "~/api/gateway/schema";
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
        <div>
          path
          <Input {...register("path")} />
        </div>
        <div>
          timeout
          <Input
            {...register("timeout", {
              valueAsNumber: true,
            })}
            type="number"
          />
        </div>
        <Controller
          control={control}
          name="methods"
          render={({ field }) => (
            <div>
              methods
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
            </div>
          )}
        />
        <Controller
          control={control}
          name="allow_anonymous"
          render={({ field }) => (
            <div className="flex items-center gap-3">
              <Switch
                defaultChecked={field.value ?? false}
                onCheckedChange={(value) => {
                  field.onChange(value);
                }}
                id={field.name}
              />
              <label htmlFor={field.name}>allow anonymous</label>
            </div>
          )}
        />
        <TargetPluginForm form={formControls} />
        <InboundPluginForm form={formControls} />
        <OutboundPluginForm form={formControls} />
        <AuthPluginForm formControls={formControls} />
      </div>
    ),
  });
};
