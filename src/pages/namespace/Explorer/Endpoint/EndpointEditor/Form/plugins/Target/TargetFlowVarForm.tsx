import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import { ModalFooter, PluginWrapper } from "../components/Modal";
import {
  TargetFlowVarFormSchema,
  TargetFlowVarFormSchemaType,
} from "../../../schema/plugins/target/targetFlowVar";

import { Fieldset } from "../../components/FormHelper";
import FilePicker from "~/componentsNext/FilePicker";
import Input from "~/design/Input";
import NamespaceSelector from "~/componentsNext/NamespaceSelector";
import { treatEmptyStringAsUndefined } from "../utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetFlowVarFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetFlowVarFormSchemaType) => void;
};

export const TargetFlowVarForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const {
    control,
    register,
    watch,
    handleSubmit,
    formState: { errors },
  } = useForm<TargetFlowVarFormSchemaType>({
    resolver: zodResolver(TargetFlowVarFormSchema),
    defaultValues: {
      type: "target-flow-var",
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.namespace"
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.workflow"
          )}
          htmlFor="workflow"
        >
          <Controller
            control={control}
            name="configuration.flow"
            render={({ field }) => (
              <FilePicker
                namespace={watch("configuration.namespace")}
                onChange={field.onChange}
                defaultPath={field.value}
                selectable={(node) => node.type === "workflow"}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.variable"
          )}
          htmlFor="variable"
        >
          <Input {...register("configuration.variable")} id="variable" />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.contentType"
          )}
          htmlFor="content-type"
        >
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            id="content-type"
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.contentTypePlaceholder"
            )}
          />
        </Fieldset>
      </PluginWrapper>
      <ModalFooter />
    </form>
  );
};
