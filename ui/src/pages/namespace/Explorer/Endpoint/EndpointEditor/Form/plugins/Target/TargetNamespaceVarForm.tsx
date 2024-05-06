import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TargetNamespaceVarFormSchema,
  TargetNamespaceVarFormSchemaType,
} from "../../../schema/plugins/target/targetNamespaceVar";

import { DisableNamespaceSelectNote } from "./utils/DisableNamespaceSelectNote";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import NamespaceSelector from "~/components/NamespaceSelector";
import NamespaceVariablePicker from "~/components/NamespaceVariablepicker";
import { PluginWrapper } from "../components/PluginSelector";
import { treatEmptyStringAsUndefined } from "~/pages/namespace/Explorer/utils";
import { useDisableNamespaceSelect } from "./utils/useDisableNamespaceSelect";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<
  TargetNamespaceVarFormSchemaType["configuration"]
>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetNamespaceVarFormSchemaType) => void;
};

export const TargetNamespaceVarForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    control,
    register,
    watch,
    setValue,
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

  const disableNamespaceSelector = useDisableNamespaceSelect();

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm} id={formId}>
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
          {disableNamespaceSelector && <DisableNamespaceSelectNote />}
          <Controller
            control={control}
            name="configuration.namespace"
            render={({ field }) => (
              <NamespaceSelector
                id="namespace"
                defaultValue={field.value}
                onValueChange={field.onChange}
                disabled={disableNamespaceSelector}
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
          <Controller
            control={control}
            name="configuration.variable"
            render={({ field }) => (
              <NamespaceVariablePicker
                defaultVariable={watch("configuration.variable")}
                namespace={watch("configuration.namespace")}
                onChange={(name, mimeType) => {
                  field.onChange(name);
                  if (mimeType) {
                    setValue("configuration.content_type", mimeType);
                  }
                }}
              />
            )}
          />
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
    </form>
  );
};
