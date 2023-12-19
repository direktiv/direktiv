import { FC, FormEvent } from "react";
import { Fieldset, ModalFooter, PluginWrapper } from "../components/Modal";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetFlowFormSchema,
  TargetFlowFormSchemaType,
} from "../../../schema/plugins/target/targetFlow";

import { Checkbox } from "~/design/Checkbox";
import Input from "~/design/Input";
import { treatEmptyStringAsUndefined } from "../utils";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetFlowFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  async: false,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetFlowFormSchemaType) => void;
};

export const TargetFlowForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<TargetFlowFormSchemaType>({
    resolver: zodResolver(TargetFlowFormSchema),
    defaultValues: {
      type: "target-flow",
      configuration: {
        ...predfinedConfig,
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.namespace"
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.workflow"
          )}
          htmlFor="workflow"
        >
          <Input {...register("configuration.flow")} id="workflow" />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.asynchronous"
          )}
          htmlFor="async"
        >
          <Checkbox
            defaultChecked={getValues("configuration.async")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.async", value);
              }
            }}
            id="async"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.contentType"
          )}
          htmlFor="content-type"
        >
          <Input
            {...register("configuration.content_type", {
              setValueAs: treatEmptyStringAsUndefined,
            })}
            placeholder={t(
              "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.contentTypePlaceholder"
            )}
            id="content-type"
          />
        </Fieldset>
      </PluginWrapper>
      <ModalFooter />
    </form>
  );
};
