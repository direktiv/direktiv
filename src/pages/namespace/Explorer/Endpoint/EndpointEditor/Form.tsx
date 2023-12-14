import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";
import { EndpointFormSchema, EndpointFormSchemaType } from "./utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import Badge from "~/design/Badge";
import { Card } from "~/design/Card";
import { Checkbox } from "~/design/Checkbox";
import { FC } from "react";
import Input from "~/design/Input";
import { Switch } from "~/design/Switch";
import { routeMethods } from "~/api/gateway/schema";
import { targetPluginTypes } from "./plugins/target";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  endpointConfig?: EndpointFormSchemaType;
  children: (args: {
    formControls: UseFormReturn<EndpointFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<EndpointFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ endpointConfig, children }) => {
  const formControls = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointFormSchema),
    defaultValues: {
      ...endpointConfig,
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
          />
        </div>
        <Controller
          control={control}
          name="methods"
          render={({ field }) => (
            <div>
              methods
              <div className="grid grid-cols-5 gap-5">
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
            <div>
              allow_anonymous
              <Switch
                defaultChecked={field.value ?? false}
                onCheckedChange={(value) => {
                  field.onChange(value);
                }}
              />
            </div>
          )}
        />

        {/* TODO: make this one new form component in an overlay with a submit */}
        <Card className="flex flex-col gap-3 p-5">
          <Controller
            control={control}
            name="plugins.target.type"
            render={({ field }) => (
              <div>
                <Select onValueChange={field.onChange} {...field}>
                  <SelectTrigger>
                    <SelectValue placeholder="please select a target plugin" />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.values(targetPluginTypes).map(
                      (targetPluginType) => (
                        <SelectItem
                          key={targetPluginType}
                          value={targetPluginType}
                        >
                          {targetPluginType}
                        </SelectItem>
                      )
                    )}
                  </SelectContent>
                </Select>
              </div>
            )}
          />
          <Controller
            control={control}
            name="plugins.target"
            render={({ field: { value: value } }) => {
              if (value.type === targetPluginTypes.instantResponse) {
                return <div>instance response flow</div>;
              }
              if (value.type === targetPluginTypes.targetFlow) {
                return <div>target flow</div>;
              }
              return <></>;
            }}
          />
        </Card>
      </div>
    ),
  });
};
