import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
} from "react-hook-form";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { ServiceFormSchema, ServiceFormSchemaType } from "../schema";

import EnvForm from "./EnvForm";
import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: ServiceFormSchemaType;
  children: (args: {
    formControls: UseFormReturn<ServiceFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<ServiceFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const formControls = useForm<ServiceFormSchemaType>({
    resolver: zodResolver(ServiceFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const fieldsInOrder = ServiceFormSchema.keyof().options;

  const values = fieldsInOrder.reduce(
    (object, key) => ({ ...object, [key]: formControls.watch(key) }),
    {}
  );

  const { register, control } = formControls;

  const sizeValues = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9"];

  return children({
    formControls,
    values,
    formMarkup: (
      <div className="flex flex-col gap-8">
        <Fieldset
          label={t("pages.explorer.service.editor.form.image")}
          htmlFor="image"
          className="grow"
        >
          <Input {...register("image")} id="image" />
        </Fieldset>

        <Fieldset
          label={t("pages.explorer.service.editor.form.scale.label")}
          htmlFor="scale"
          className="grow"
        >
          <Select
            value={formControls.getValues("scale")?.toString()}
            onValueChange={(value) =>
              formControls.setValue("scale", Number(value))
            }
          >
            <SelectTrigger>
              <SelectValue
                placeholder={t(
                  "pages.explorer.service.editor.form.scale.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectLabel>Size</SelectLabel>
                {sizeValues.map((value, index) => (
                  <SelectItem key={index} value={value}>
                    {value}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </Fieldset>
        <Fieldset
          label={t("pages.explorer.service.editor.form.size.label")}
          htmlFor="size"
          className="grow"
        >
          <Select
            value={formControls.getValues("size")}
            onValueChange={(value) => formControls.setValue("size", value)}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={t(
                  "pages.explorer.service.editor.form.size.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectLabel>Size</SelectLabel>
                <SelectItem value="small">small</SelectItem>
                <SelectItem value="medium">medium</SelectItem>
                <SelectItem value="large">large</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </Fieldset>

        <Fieldset
          label={t("pages.explorer.service.editor.form.cmd")}
          htmlFor="cmd"
          className="grow"
        >
          <Input {...register("cmd")} id="cmd" />
        </Fieldset>

        <Fieldset
          label={t("pages.explorer.service.editor.form.envs.label")}
          htmlFor="size"
          className="flex grow"
        >
          <Controller
            control={control}
            name="envs"
            render={({ field }) => (
              <EnvForm
                defaultValue={field.value || []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
      </div>
    ),
  });
};
