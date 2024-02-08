import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
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
import {
  ServiceFormSchema,
  ServiceFormSchemaType,
  scaleOptions,
} from "../schema";

import { EnvsArrayInput } from "./EnvsArrayInput";
import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { PatchesForm } from "./Patches";
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

  const watchedValues = useWatch({
    control: formControls.control,
  });

  const values = fieldsInOrder.reduce(
    (object, key) => ({ ...object, [key]: watchedValues[key] }),
    {}
  );

  const { register, control } = formControls;

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

        <div className="grid grid-cols-2 gap-4">
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
              <SelectTrigger variant="outline" id="scale">
                <SelectValue
                  placeholder={t(
                    "pages.explorer.service.editor.form.scale.placeholder"
                  )}
                />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>Size</SelectLabel>
                  {scaleOptions.map((value, index) => (
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
              <SelectTrigger variant="outline" id="size">
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
        </div>

        <Fieldset
          label={t("pages.explorer.service.editor.form.cmd")}
          htmlFor="cmd"
          className="grow"
        >
          <Input {...register("cmd")} id="cmd" />
        </Fieldset>

        <PatchesForm form={formControls} />

        <Fieldset
          label={t("pages.explorer.service.editor.form.envs.label")}
          htmlFor="size"
          className="flex grow"
        >
          <Controller
            control={control}
            name="envs"
            render={({ field }) => <EnvsArrayInput field={field} />}
          />
        </Fieldset>
      </div>
    ),
  });
};
