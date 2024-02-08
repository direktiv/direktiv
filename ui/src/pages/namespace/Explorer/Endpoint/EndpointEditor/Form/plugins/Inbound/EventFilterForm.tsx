import { Controller, useForm } from "react-hook-form";
import {
  EventFilterFormSchema,
  EventFilterFormSchemaType,
} from "../../../schema/plugins/inbound/eventFilter";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { Card } from "~/design/Card";
import { Checkbox } from "~/design/Checkbox";
import Editor from "~/design/Editor";
import { Fieldset } from "~/components/Form/Fieldset";
import { PluginWrapper } from "../components/PluginSelector";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<EventFilterFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
  allow_non_events: false,
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: EventFilterFormSchemaType) => void;
};

export const EventFilterForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
    control,
  } = useForm<EventFilterFormSchemaType>({
    resolver: zodResolver(EventFilterFormSchema),
    defaultValues: {
      type: "event-filter",
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

  const theme = useTheme();

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
            "pages.explorer.endpoint.editor.form.plugins.inbound.eventFilter.script"
          )}
        >
          <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
            <Controller
              control={control}
              name="configuration.script"
              render={({ field }) => (
                <Editor
                  theme={theme ?? undefined}
                  language="javascript"
                  value={field.value}
                  onChange={field.onChange}
                />
              )}
            />
          </Card>
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.eventFilter.allow_non_events"
          )}
          htmlFor="allow_non_events"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.allow_non_events")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.allow_non_events", value);
              }
            }}
            id="allow_non_events"
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
