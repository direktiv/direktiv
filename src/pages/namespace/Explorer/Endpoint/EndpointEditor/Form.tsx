import Alert from "~/design/Alert";
import { EndpointFormSchemaType } from "../utils";
import { FC } from "react";
import { useTranslation } from "react-i18next";

type FormProps = {
  endpointConfig?: EndpointFormSchemaType;
};

export const Form: FC<FormProps> = ({ endpointConfig }) => {
  const { t } = useTranslation();
  if (!endpointConfig) {
    return (
      <Alert variant="error">
        {t("pages.explorer.endpoint.editor.form.serialisationError")}
      </Alert>
    );
  }

  return <div>FORM</div>;
};
