import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  RequestConvertFormSchema,
  RequestConvertFormSchemaType,
} from "../../../schema/plugins/inbound/requestConvert";

import { Checkbox } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { PluginWrapper } from "../components/Modal";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<RequestConvertFormSchemaType["configuration"]>;

const predefinedConfig: OptionalConfig = {
  omit_body: false,
  omit_headers: false,
  omit_consumer: false,
  omit_queries: false,
};

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: RequestConvertFormSchemaType) => void;
};

export const RequestConvertForm: FC<FormProps> = ({
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
  } = useForm<RequestConvertFormSchemaType>({
    resolver: zodResolver(RequestConvertFormSchema),
    defaultValues: {
      type: "request-convert",
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
            "pages.explorer.endpoint.editor.form.plugins.inbound.requestConvert.omitHeaders"
          )}
          htmlFor="omit-headers"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.omit_headers")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_headers", value);
              }
            }}
            id="omit-headers"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.requestConvert.omitQueries"
          )}
          htmlFor="omit-queries"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.omit_queries")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_queries", value);
              }
            }}
            id="omit-queries"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.requestConvert.omitBody"
          )}
          htmlFor="omit-body"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.omit_body")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_body", value);
              }
            }}
            id="omit-body"
          />
        </Fieldset>
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.requestConvert.omitConsumer"
          )}
          htmlFor="omit-consumer"
          horizontal
        >
          <Checkbox
            defaultChecked={getValues("configuration.omit_consumer")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.omit_consumer", value);
              }
            }}
            id="omit-consumer"
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
