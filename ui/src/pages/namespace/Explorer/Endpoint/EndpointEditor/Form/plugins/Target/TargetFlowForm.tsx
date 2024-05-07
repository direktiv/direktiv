import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TargetFlowFormSchema,
  TargetFlowFormSchemaType,
} from "../../../schema/plugins/target/targetFlow";

import { Checkbox } from "~/design/Checkbox";
import { DisableNamespaceSelectNote } from "./utils/DisableNamespaceSelectNote";
import { Fieldset } from "~/components/Form/Fieldset";
import FilePicker from "~/components/FilePicker";
import Input from "~/design/Input";
import NamespaceSelector from "~/components/NamespaceSelector";
import { PluginWrapper } from "../components/PluginSelector";
import { treatEmptyStringAsUndefined } from "~/pages/namespace/Explorer/utils";
import { useIsSystemNamespace } from "./utils/useIsSystemNamespace";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetFlowFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
  async: false,
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetFlowFormSchemaType) => void;
};

export const TargetFlowForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    setValue,
    getValues,
    watch,
    control,
    formState: { errors },
  } = useForm<TargetFlowFormSchemaType>({
    resolver: zodResolver(TargetFlowFormSchema),
    defaultValues: {
      type: "target-flow",
      configuration: {
        ...predefinedConfig,
        ...defaultConfig,
      },
    },
  });

  const disableNamespaceSelect = useIsSystemNamespace();

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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.namespace"
          )}
          htmlFor="namespace"
        >
          {disableNamespaceSelect && <DisableNamespaceSelectNote />}
          <Controller
            control={control}
            name="configuration.namespace"
            render={({ field }) => (
              <NamespaceSelector
                id="namespace"
                defaultValue={field.value}
                onValueChange={field.onChange}
                disabled={disableNamespaceSelect}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.workflow"
          )}
        >
          <Controller
            control={control}
            name="configuration.flow"
            render={({ field }) => (
              <FilePicker
                namespace={watch("configuration.namespace")}
                onChange={field.onChange}
                defaultPath={field.value}
                selectable={(file) => file.type === "workflow"}
              />
            )}
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.asynchronous"
          )}
          htmlFor="async"
          horizontal
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
    </form>
  );
};
