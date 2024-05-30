import { QueryErrorType, getMessageFromApiError } from "~/api/errorHandling";

import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

const ApiError = ({
  error,
  label,
  className,
}: {
  error: QueryErrorType;
  label?: string;
  className?: string;
}) => {
  const { t } = useTranslation();
  const errorMessage = getMessageFromApiError(error);
  const errorLabel = label || t("components.apiError.label");

  return (
    <Alert variant="error" className={className}>
      {errorLabel} {errorMessage}
    </Alert>
  );
};

export default ApiError;
