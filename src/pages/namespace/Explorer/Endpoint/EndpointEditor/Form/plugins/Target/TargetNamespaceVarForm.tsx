import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import { Fieldset, ModalFooter, PluginWrapper } from "../components/Modal";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetNamespaceVarFormSchema,
  TargetNamespaceVarFormSchemaType,
} from "../../../schema/plugins/target/targetNamespaceVar";

import Input from "~/design/Input";
import NamespaceSelector from "~/componentsNext/NamespaceSelector";
import { treatEmptyStringAsUndefined } from "../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<
  TargetNamespaceVarFormSchemaType["configuration"]
>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetNamespaceVarFormSchemaType) => void;
};

export const TargetNamespaceVarForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TargetNamespaceVarFormSchemaType>({
    resolver: zodResolver(TargetNamespaceVarFormSchema),
    defaultValues: {
      type: "target-namespace-var",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm}>
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetNamespaceVariable.namespace"
          )}
          htmlFor="namespace"
        >
          <Controller
            control={control}
            name="configuration.namespace"
            render={({ field }) => (
              <NamespaceSelector
                id="namespace"
                defaultValue={field.value}
                onValueChange={field.onChange}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetNamespaceVariable.variable"
          )}
          htmlFor="variable"
        >
          <Input {...register("configuration.variable")} id="variable" />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetNamespaceVariable.contentType"
          )}
          htmlFor="content-type"
        >
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.targetNamespaceVariable.contentTypePlaceholder"
            )}
            id="content-type"
          />
        </Fieldset>
      </PluginWrapper>
      <ModalFooter />
    </form>
  );
};
