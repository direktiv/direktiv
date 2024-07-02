import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TargetEventFormSchema,
  TargetEventFormSchemaType,
} from "../../../schema/plugins/target/targetEvent";

import { Command } from "~/design/Command";
import { DisableNamespaceSelectNote } from "./utils/DisableNamespaceSelectNote";
import { Fieldset } from "~/components/Form/Fieldset";
import { NamespaceSelectorList } from "~/components/Breadcrumb/NamespaceSelectorList";
import { PluginWrapper } from "../components/PluginSelector";
import { useIsSystemNamespace } from "./utils/useIsSystemNamespace";
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
  const isSystemNamespace = useIsSystemNamespace();

  const predefinedConfig = isSystemNamespace ? { namespaces: [] } : {};

  const {
    handleSubmit,
    control,
    setValue,
    formState: { errors },
  } = useForm<TargetEventFormSchemaType>({
    resolver: zodResolver(TargetEventFormSchema),
    defaultValues: {
      type: "target-event",
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
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}

        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetEvent.namespace"
          )}
          htmlFor="namespace"
        >
          {!isSystemNamespace ? (
            <DisableNamespaceSelectNote />
          ) : (
            <Controller
              control={control}
              name="configuration.namespaces"
              render={({ field }) => (
                <Command id="namespace">
                  <NamespaceSelectorList
                    onSelectNamespace={(value) => {
                      if (field.value.includes(value)) {
                        return setValue(
                          "configuration.namespaces",
                          field.value.filter((item) => item !== value)
                        );
                      }
                      setValue("configuration.namespaces", [
                        ...field.value,
                        value,
                      ]);
                    }}
                    isMulti={true}
                    selectedValues={field.value || []}
                  />
                </Command>
              )}
            />
          )}
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
