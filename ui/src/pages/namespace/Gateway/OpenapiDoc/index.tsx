import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { RapiDoc } from "~/design/RapiDoc";
import { Save } from "lucide-react";
import { useDocumentation } from "~/api/gateway/query/getDocumentation";
import { useState } from "react";
import { useTranslation } from "react-i18next";

interface Spec {
  openapi: string;
  info: {
    title: string;
    version: string;
  };
  paths: Record<string, unknown>;
}

interface DocumentationInfo {
  spec: Spec;
  errors: string[];
}

const OpenapiDocPage: React.FC = () => {
  const { data } = useDocumentation();
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const [copyError, setCopyError] = useState<string | null>(null);

  const info = data?.data as DocumentationInfo | undefined;
  const { spec, errors } = info || { spec: null, errors: [] };

  const hasErrors = errors && errors.length > 0;
  const hasSpec = spec && spec.paths && Object.keys(spec.paths).length > 0;

  const handleCopy = async () => {
    if (!spec) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(spec, null, 2));
      setCopied(true);
      setCopyError(null);
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      setCopyError(
        error instanceof Error
          ? error.message
          : t("pages.gateway.documentation.error")
      );
      setTimeout(() => setCopyError(null), 4000);
    }
  };

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <Card className="w-full">
        <div className="flex flex-col">
          {hasErrors && (
            <Alert variant="error">
              <pre>{JSON.stringify(errors, null, 2)}</pre>
            </Alert>
          )}
          {hasSpec && <RapiDoc spec={spec} className="h-[75vh] my-1" />}
          {!hasSpec && !hasErrors && (
            <Alert variant="info">
              <p>{t("pages.gateway.documentation.noDocumentation")}</p>
            </Alert>
          )}
        </div>
      </Card>
      {copyError && (
        <Alert variant="error">
          <p>{copyError}</p>
        </Alert>
      )}
      <div className="flex justify-end">
        <Button variant="outline" onClick={handleCopy}>
          <Save />
          {copied
            ? t("pages.gateway.documentation.copied")
            : t("pages.gateway.documentation.copyAPI")}
        </Button>
      </div>
    </div>
  );
};

export default OpenapiDocPage;
