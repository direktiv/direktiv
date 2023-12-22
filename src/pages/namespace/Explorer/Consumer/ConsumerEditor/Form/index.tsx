import { ConsumerFormSchema, ConsumerFormSchemaType } from "../schema";
import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";

import { ArrayInput } from "../../../components/ArrayInput";
import { FC } from "react";
import { Fieldset } from "../../../components/Fieldset";
import Input from "~/design/Input";
import { treatEmptyStringAsUndefined } from "../../../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: ConsumerFormSchemaType;
  children: (args: {
    formControls: UseFormReturn<ConsumerFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<ConsumerFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const formControls = useForm<ConsumerFormSchemaType>({
    resolver: zodResolver(ConsumerFormSchema),
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
            />
          </Fieldset>
        </div>
        <Fieldset
          label={t("pages.explorer.consumer.editor.form.api_key")}
          htmlFor="api-key"
        >
          <Input
            {...register("api_key", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            id="api-key"
            type="password"
          />
        </Fieldset>
        <Fieldset label={t("pages.explorer.consumer.editor.form.groups")}>
          <Controller
            control={control}
            name="groups"
            render={({ field }) => (
              <ArrayInput
                placeholder={t(
                  "pages.explorer.consumer.editor.form.groupsPlaceholder"
                )}
                defaultValue={field.value ?? []}
                onChange={(changedValue) => {
                  field.onChange(changedValue);
                }}
              />
            )}
          />
        </Fieldset>
        <Fieldset label={t("pages.explorer.consumer.editor.form.tags")}>
          <Controller
            control={control}
            name="tags"
            render={({ field }) => (
              <ArrayInput
                placeholder={t(
                  "pages.explorer.consumer.editor.form.tagsPlaceholder"
                )}
                defaultValue={field.value ?? []}
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
