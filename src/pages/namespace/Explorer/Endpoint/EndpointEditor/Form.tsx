import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";
import { EndpointFormSchema, EndpointFormSchemaType } from "../utils";

import { FC } from "react";
import Input from "~/design/Input";
import { Switch } from "~/design/Switch";
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
      <div className="flex flex-col gap-3">
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
      </div>
    ),
  });
};
