import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";

const SuspenseBoundary = ({ children }: { children: React.ReactNode }) => {
  const { t } = useTranslation();
  return (
    <Suspense fallback={<Loading />}>
      <ErrorBoundary
        fallbackRender={({ error, resetErrorBoundary }) => (
          <ParsingError
            title={t("direktivPage.error.genericError")}
            resetErrorBoundary={resetErrorBoundary}
          >
            {error.message}
          </ParsingError>
        )}
      >
        {children}
      </ErrorBoundary>
    </Suspense>
  );
};

export { SuspenseBoundary as BlockSuspenseBoundary };
