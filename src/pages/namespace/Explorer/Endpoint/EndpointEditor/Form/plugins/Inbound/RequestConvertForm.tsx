import { FC, FormEvent } from "react";
import { Fieldset, ModalFooter, PluginWrapper } from "../components/Modal";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  RequestConvertFormSchema,
  RequestConvertFormSchemaType,
} from "../../../schema/plugins/inbound/requestConvert";

import { Checkbox } from "~/design/Checkbox";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<RequestConvertFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  omit_body: false,
  omit_headers: false,
  omit_consumer: false,
  omit_queries: false,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: RequestConvertFormSchemaType) => void;
};

export const RequestConvertForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
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
            "pages.explorer.endpoint.editor.form.plugins.inbound.requestConvert.omitHeaders"
          )}
          htmlFor="omit-headers"
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
      <ModalFooter />
    </form>
  );
};
