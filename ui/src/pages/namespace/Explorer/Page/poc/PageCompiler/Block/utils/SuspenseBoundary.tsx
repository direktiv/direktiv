import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { Suspense } from "react";
import { usePage } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";

const SuspenseBoundary = ({ children }: { children: React.ReactNode }) => {
  const { t } = useTranslation();
  const page = usePage();
  return (
    <Suspense fallback={<Loading />}>
      <ErrorBoundary
        fallbackRender={({ error, resetErrorBoundary }) => (
          <ParsingError
            title={t("direktivPage.error.genericError")}
            resetErrorBoundary={resetErrorBoundary}
            page={page}
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
