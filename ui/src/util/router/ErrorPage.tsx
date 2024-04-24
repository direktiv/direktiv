import { ArrowLeft, Home, RefreshCw } from "lucide-react";
import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom";

import Button from "~/design/Button";
import Logo from "~/components/Logo";
import { isApiErrorSchema } from "~/api/errorHandling";
import { twMergeClsx } from "../helpers";
import { useTranslation } from "react-i18next";

type ErrorPageProps = {
  className?: string;
};
const ErrorPage = ({ className }: ErrorPageProps) => {
  const { t } = useTranslation();
  const error = useRouteError();

  let errorTitle = t("pages.error.status");
  let errorMessage = t("pages.error.message");

  if (isRouteErrorResponse(error)) {
    errorTitle = `${error.status}`;
    errorMessage = error.statusText;
  }

  if (isApiErrorSchema(error) && error.response.status === 404) {
    errorTitle = `${error.response.status}`;
    errorMessage = t("pages.error.notFound");
  }

  return (
    <main
      className={twMergeClsx(
        "flex h-screen w-full flex-col items-center justify-center",
        className
      )}
    >
      <div className="flex max-w-xs flex-col items-center gap-3">
        <Logo />
        <div className="text-4xl font-bold" data-testid="error-title">
          {errorTitle}
        </div>
        <div data-testid="error-message" className="text-center">
          {errorMessage}
        </div>
        <div className="grid w-full grid-cols-2 gap-3">
          <Button
            variant="outline"
            onClick={() => window.history.back()}
            data-testid="error-back-btn"
          >
            <ArrowLeft />
            {t("pages.error.goBack")}
          </Button>
          <Button
            variant="outline"
            onClick={() => location.reload()}
            data-testid="error-reload-btn"
          >
            <RefreshCw />
            {t("pages.error.reload")}
          </Button>
          <Button variant="primary" asChild isAnchor className="col-span-2">
            <Link to="/">
              <Home />
              {t("pages.error.goHome")}
            </Link>
          </Button>
        </div>
      </div>
    </main>
  );
};

export default ErrorPage;
