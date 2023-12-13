import { EndpointFormSchema, EndpointFormSchemaType } from "../utils";

import Alert from "~/design/Alert";
import { Card } from "@tremor/react";
import { Checkbox } from "~/design/Checkbox";
import { FC } from "react";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  endpointConfig?: EndpointFormSchemaType;
};

export const Form: FC<FormProps> = ({ endpointConfig }) => {
  const { t } = useTranslation();

  const {
    register,
    handleSubmit,
    setValue,
    watch,

    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointFormSchema),
    defaultValues: {
      ...endpointConfig,
    },
  });

  if (!endpointConfig) {
    return (
      <Alert variant="error">
        {t("pages.explorer.endpoint.editor.form.serialisationError")}
      </Alert>
    );
  }

  return (
    <Card>
      <Input {...register("path")} />
      {watch("direktiv_api")}
      <hr />
      {watch("path")}
    </Card>
  );
};
