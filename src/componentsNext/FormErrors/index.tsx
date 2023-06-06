import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

type errorsType = Record<string, { message?: string }>;

const FormErrors = ({ errors }: { errors: errorsType }) => {
  const { t } = useTranslation();
  const entries = Object.entries(errors);

  return entries.length ? (
    <Alert variant="error" className="mb-5">
      <ul>
        {entries.map(([key, value]) => (
          <li key={key}>{`${key}: ${
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
