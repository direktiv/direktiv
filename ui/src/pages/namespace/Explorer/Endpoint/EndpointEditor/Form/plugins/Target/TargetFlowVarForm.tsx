import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TargetFlowVarFormSchema,
  TargetFlowVarFormSchemaType,
} from "../../../schema/plugins/target/targetFlowVar";

import { Fieldset } from "~/components/Form/Fieldset";
import FilePicker from "~/components/FilePicker";
import Input from "~/design/Input";
import NamespaceSelector from "~/components/NamespaceSelector";
import { PluginWrapper } from "../components/Modal";
import WorkflowVariablePicker from "~/components/WorkflowVariablepicker";
import { treatEmptyStringAsUndefined } from "~/pages/namespace/Explorer/utils";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetFlowVarFormSchemaType["configuration"]>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetFlowVarFormSchemaType) => void;
};

export const TargetFlowVarForm: FC<FormProps> = ({
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
          <Controller
            control={control}
            name="configuration.variable"
            render={({ field }) => (
              <WorkflowVariablePicker
                namespace={watch("configuration.namespace")}
                workflowPath={watch("configuration.flow")}
                defaultVariable={watch("configuration.variable")}
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.contentType"
          )}
          htmlFor="content-type"
        >
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.contentTypePlaceholder"
            )}
            id="content-type"
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
