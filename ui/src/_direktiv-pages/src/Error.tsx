import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

const ErrorMessage = ({ error }: { error: string }) => {
  const { t } = useTranslation();

  return (
    <div className="flex h-screen items-center justify-center">
      <Alert variant="error" className="max-w-lg">
        <div className="font-bold">{t("direktivPage.error.genericError")}</div>
        {error}
      </Alert>
    </div>
  );
};

export default ErrorMessage;
