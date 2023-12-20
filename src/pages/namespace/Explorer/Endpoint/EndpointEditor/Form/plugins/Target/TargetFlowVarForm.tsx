import { FC, FormEvent } from "react";
import { Fieldset, ModalFooter, PluginWrapper } from "../components/Modal";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetFlowVarFormSchema,
  TargetFlowVarFormSchemaType,
} from "../../../schema/plugins/target/targetFlowVar";

import Input from "~/design/Input";
import { treatEmptyStringAsUndefined } from "../utils";
import { useForm } from "react-hook-form";
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
    register,
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
          <Input
            {...register("configuration.namespace", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            id="namespace"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlowVar.workflow"
          )}
          htmlFor="workflow"
        >
          <Input {...register("configuration.flow")} id="workflow" />
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
