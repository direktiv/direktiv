import { QueryErrorType, getMessageFromApiError } from "~/api/errorHandling";

import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

const ApiError = ({
  error,
  className,
}: {
  error: QueryErrorType;
  className?: string;
}) => {
  const { t } = useTranslation();
  const errorMessage = getMessageFromApiError(error);

  return (
    <Alert variant="error" className={className}>
      {errorMessage
        ? t("components.apiError.withMessage", { message: errorMessage })
        : t("components.apiError.onlyLabel")}
    </Alert>
  );
};

export default ApiError;
