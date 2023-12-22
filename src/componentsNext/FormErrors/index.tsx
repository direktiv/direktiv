import { ComponentProps, FC } from "react";

import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

export type errorsType = Record<string, { message?: string }>;

type FormErrorsProps = ComponentProps<typeof Alert> & { errors: errorsType };

const FormErrors: FC<FormErrorsProps> = ({ errors, ...props }) => {
  const { t } = useTranslation();
  const entries = Object.entries(errors);
  return entries.length ? (
    <Alert variant="error" {...props} data-testid="form-errors">
      <ul>
        {entries.map(([key, value], index) => (
          // note: key might also be an empty string
          <li key={index}>{`${key && `${key} :`} ${
            value.message || t("components.formErrors.fieldInvalid")
          }`}</li>
        ))}
      </ul>
    </Alert>
  ) : (
    <></>
  );
};

export default FormErrors;
