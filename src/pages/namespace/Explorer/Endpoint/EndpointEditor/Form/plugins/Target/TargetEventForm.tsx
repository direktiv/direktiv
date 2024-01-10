import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  TargetEventFormSchema,
  TargetEventFormSchemaType,
} from "../../../schema/plugins/target/targetEvent";

import { Fieldset } from "~/pages/namespace/Explorer/components/Fieldset";
import NamespaceSelector from "~/componentsNext/NamespaceSelector";
import { PluginWrapper } from "../components/Modal";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetEventFormSchemaType["configuration"]>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetEventFormSchemaType) => void;
};

export const TargetEventForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<TargetEventFormSchemaType>({
    resolver: zodResolver(TargetEventFormSchema),
    defaultValues: {
      type: "target-event",
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
            "pages.explorer.endpoint.editor.form.plugins.target.targetFlow.namespace"
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
      </PluginWrapper>
    </form>
  );
};
