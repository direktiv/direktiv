import { ConsumerFormSchema, ConsumerFormSchemaType } from "../schema";
import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";

import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { TagsGroupsArrayForm } from "./TagsGroupsArrayForm";
import { treatEmptyStringAsUndefined } from "../../../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: ConsumerFormSchemaType;
  children: (args: {
    form: UseFormReturn<ConsumerFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<ConsumerFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const form = useForm<ConsumerFormSchemaType>({
    resolver: zodResolver(ConsumerFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const fieldsInOrder = ConsumerFormSchema.keyof().options;

  const watchedValues = useWatch({
    control: form.control,
  });

  const values = fieldsInOrder.reduce(
    (object, key) => ({ ...object, [key]: watchedValues[key] }),
    {}
  );

  const { register, control } = form;

  return children({
    form,
    values,
    formMarkup: (
      <div className="flex flex-col gap-6">
        <div className="flex gap-3">
          <Fieldset
            label={t("pages.explorer.consumer.editor.form.username")}
            htmlFor="username"
            className="grow"
          >
            <Input {...register("username")} id="username" />
          </Fieldset>
          <Fieldset
            label={t("pages.explorer.consumer.editor.form.password")}
            htmlFor="password"
            className="grow"
          >
            <Input
              {...register("password", {
                setValueAs: treatEmptyStringAsUndefined,
              })}
              id="password"
              type="password"
              autoComplete="new-password"
            />
          </Fieldset>
        </div>
        <Fieldset
          label={t("pages.explorer.consumer.editor.form.apiKey")}
          htmlFor="api-key"
        >
          <Input
            {...register("api_key", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            id="api-key"
            type="password"
            autoComplete="new-password"
          />
        </Fieldset>
        <Fieldset label={t("pages.explorer.consumer.editor.form.groups")}>
          <Controller
            control={control}
            name="groups"
            render={({ field }) => <TagsGroupsArrayForm field={field} />}
          />
        </Fieldset>
        <Fieldset label={t("pages.explorer.consumer.editor.form.tags")}>
          <Controller
            control={control}
            name="tags"
            render={({ field }) => <TagsGroupsArrayForm field={field} />}
          />
        </Fieldset>
      </div>
    ),
  });
};
