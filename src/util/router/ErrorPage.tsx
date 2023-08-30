import { ArrowLeft, Home, RefreshCcw } from "lucide-react";
import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom";

import Button from "~/design/Button";
import Logo from "~/design/Logo";
import { useTranslation } from "react-i18next";

const ErrorPage = () => {
  const { t } = useTranslation();
  const error = useRouteError();

  let errorTitle = t("pages.error.status");
  let errorMessage = t("pages.error.message");

  if (isRouteErrorResponse(error)) {
    errorTitle = `${error.status}`;
    errorMessage = error.statusText;
  }

  return (
    <main className="flex h-screen w-full flex-col items-center justify-center gap-3">
      <Logo />
      <div className="text-4xl font-bold">{errorTitle}</div>
      <div>{errorMessage}</div>
      <div className="grid grid-cols-2 gap-3">
        <Button variant="outline" onClick={() => window.history.back()}>
          <ArrowLeft />
          {t("pages.error.goBack")}
        </Button>
        <Button variant="outline" onClick={() => location.reload()}>
          <RefreshCcw />
          {t("pages.error.reload")}
        </Button>
        <Button variant="primary" asChild className="col-span-2">
          <Link to="/">
            <Home />
            {t("pages.error.goHome")}
          </Link>
        </Button>
      </div>
    </main>
  );
};

export default ErrorPage;
