import { QueryErrorType, getMessageFromApiError } from "~/api/errorHandling";

import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

const ApiError = ({
  error,
  label,
}: {
  error: QueryErrorType;
  label?: string;
}) => {
  const { t } = useTranslation();
  const errorMessage = getMessageFromApiError(error);
  const errorLabel = label || t("components.apiError.label");

  return (
    <Alert variant="error">
      {errorLabel} {errorMessage}
    </Alert>
  );
};

export default ApiError;
