import { ConsumerFormSchema, ConsumerFormSchemaType } from "../schema";
import {
  Controller,
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";

import { ArrayInput } from "~/components/Form/ArrayInput";
import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
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

  const fieldsInOrder = ConsumerFormSchema.keyof().options;

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
              <div className="grid gap-5 sm:grid-cols-2">
                <ArrayInput
                  defaultValue={field.value || []}
                  onChange={(changedValue) => {
                    field.onChange(changedValue);
                  }}
                  emptyItem=""
                  itemIsValid={(item) => item !== ""}
                  renderItem={({
                    value,
                    setValue,
                    onChange,
                    handleKeyDown,
                  }) => (
                    <Input
                      placeholder={t(
                        "pages.explorer.consumer.editor.form.groupsPlaceholder"
                      )}
                      value={value}
                      onKeyDown={handleKeyDown}
                      onChange={(e) => {
                        const newValue = e.target.value;
                        setValue(newValue);
                        onChange(newValue);
                      }}
                    />
                  )}
                />
              </div>
            )}
          />
        </Fieldset>
        <Fieldset label={t("pages.explorer.consumer.editor.form.tags")}>
          <Controller
            control={control}
            name="tags"
            render={({ field }) => (
              <div className="grid gap-5 sm:grid-cols-2">
                <ArrayInput
                  defaultValue={field.value || []}
                  onChange={(changedValue) => {
                    field.onChange(changedValue);
                  }}
                  emptyItem=""
                  itemIsValid={(item) => item !== ""}
                  renderItem={({
                    value,
                    setValue,
                    onChange,
                    handleKeyDown,
                  }) => (
                    <Input
                      placeholder={t(
                        "pages.explorer.consumer.editor.form.tagsPlaceholder"
                      )}
                      value={value}
                      onKeyDown={handleKeyDown}
                      onChange={(e) => {
                        const newValue = e.target.value;
                        setValue(newValue);
                        onChange(newValue);
                      }}
                    />
                  )}
                />
              </div>
            )}
          />
        </Fieldset>
      </div>
    ),
  });
};
