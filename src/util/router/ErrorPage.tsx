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
    <main className="flex h-screen w-full flex-col items-center justify-center gap-8">
      <Logo />
      <div className="flex flex-col items-center gap-3">
        <div className="text-4xl font-bold">{errorTitle}</div>
        <div>{errorMessage}</div>
        <div className="flex gap-3">
          <Button variant="outline" onClick={() => window.history.back()}>
            {t("pages.error.goBack")}
          </Button>
          <Button variant="primary" asChild>
            <Link to="/">{t("pages.error.goHome")}</Link>
          </Button>
        </div>
      </div>
    </main>
  );
};

export default ErrorPage;
