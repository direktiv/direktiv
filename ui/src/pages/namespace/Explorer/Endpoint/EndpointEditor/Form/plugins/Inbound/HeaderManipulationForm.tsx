import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  HeaderManipulationFormSchema,
  HeaderManipulationFormSchemaType,
} from "../../../schema/plugins/inbound/headerManipulation";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { ObjArrayInput } from "~/components/Form/ObjArrayInput";
import { PluginWrapper } from "../components/Modal";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<
  HeaderManipulationFormSchemaType["configuration"]
>;

const predefinedConfig: OptionalConfig = {
  headers_to_add: [],
  headers_to_modify: [],
  headers_to_remove: [],
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: HeaderManipulationFormSchemaType) => void;
};

export const HeaderManipulationForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<HeaderManipulationFormSchemaType>({
    resolver: zodResolver(HeaderManipulationFormSchema),
    defaultValues: {
      type: "header-manipulation",
      configuration: {
        ...predefinedConfig,
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm} id={formId}>
      {errors?.configuration && (
        <FormErrors
          errors={errors?.configuration as errorsType}
          className="mb-5"
        />
      )}
      <PluginWrapper>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.headers_to_add"
          )}
        >
          <Controller
            control={control}
            name="configuration.headers_to_add"
            render={({ field }) => (
              <div className="grid gap-5">
                <ObjArrayInput
                  defaultValue={field.value || []}
                  onChange={(changedValue) => {
                    field.onChange(changedValue);
                  }}
                  emptyItem={{ name: "", value: "" }}
                  itemIsValid={(item) => !!(item?.name && item?.value)}
                  renderItem={({
                    state,
                    setState,
                    onChange,
                    handleKeyDown,
                  }) => (
                    <>
                      {Object.entries(state).map(([key, value]) => {
                        const typedKey = key as keyof typeof state;
                        return (
                          <Input
                            placeholder={t(
                              `pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.${typedKey}Placeholder`
                            )}
                            key={key}
                            value={value}
                            onKeyDown={handleKeyDown}
                            onChange={(e) => {
                              const newVal = { [key]: e.target.value };
                              const newState = { ...state, ...newVal };
                              setState(newState);
                              onChange(newState);
                            }}
                          />
                        );
                      })}
                    </>
                  )}
                />
              </div>
            )}
          />
        </Fieldset>

        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.headers_to_modify"
          )}
        >
          <Controller
            control={control}
            name="configuration.headers_to_modify"
            render={({ field }) => (
              <div className="grid gap-5">
                <ObjArrayInput
                  defaultValue={field.value || []}
                  onChange={(changedValue) => {
                    field.onChange(changedValue);
                  }}
                  emptyItem={{
                    name: "",
                    value: "",
                  }}
                  itemIsValid={(item) => !!(item?.name && item?.value)}
                  renderItem={({
                    state,
                    setState,
                    onChange,
                    handleKeyDown,
                  }) => (
                    <>
                      {Object.entries(state).map(([key, value]) => {
                        const typedKey = key as keyof typeof state;
                        return (
                          <Input
                            placeholder={t(
                              `pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.${typedKey}Placeholder`
                            )}
                            key={key}
                            value={value}
                            onKeyDown={handleKeyDown}
                            onChange={(e) => {
                              const newVal = { [key]: e.target.value };
                              const newState = { ...state, ...newVal };
                              setState(newState);
                              onChange(newState);
                            }}
                          />
                        );
                      })}
                    </>
                  )}
                />
              </div>
            )}
          />
        </Fieldset>

        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.headers_to_remove"
          )}
        >
          <Controller
            control={control}
            name="configuration.headers_to_remove"
            render={({ field }) => (
              <div className="grid gap-5">
                <ObjArrayInput
                  defaultValue={field.value || []}
                  onChange={(changedValue) => {
                    field.onChange(changedValue);
                  }}
                  emptyItem={{
                    name: "",
                  }}
                  itemIsValid={(item) => !!item?.name}
                  renderItem={({
                    state,
                    setState,
                    onChange,
                    handleKeyDown,
                  }) => (
                    <>
                      {Object.entries(state).map(([key, value]) => {
                        const typedKey = key as keyof typeof state;
                        return (
                          <Input
                            placeholder={t(
                              `pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.${typedKey}Placeholder`
                            )}
                            key={key}
                            value={value}
                            onKeyDown={handleKeyDown}
                            onChange={(e) => {
                              const newVal = { [key]: e.target.value };
                              const newState = { ...state, ...newVal };
                              setState(newState);
                              onChange(newState);
                            }}
                          />
                        );
                      })}
                    </>
                  )}
                />
              </div>
            )}
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
