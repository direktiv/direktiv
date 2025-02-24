import Alert from "~/design/Alert";
import { Card } from "~/design/Card";
import { RapiDoc } from "~/design/RapiDoc";
import { useDocumentation } from "~/api/gateway/query/getDocumentation";
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

  const info = data?.data as DocumentationInfo | undefined;
  const { spec, errors } = info || { Spec: null, errors: [] };

  const hasErrors = errors && errors.length > 0;
  const hasSpec = spec && spec.paths && Object.keys(spec.paths).length > 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row w-full">
        <Card className="size-full">
          <div className="flex flex-col">
            {hasErrors && (
              <Alert variant="error">
                <pre>{JSON.stringify(errors, null, 2)}</pre>
              </Alert>
            )}
            {hasSpec && <RapiDoc spec={spec} className="h-[80vh] my-1" />}
            {!hasSpec && !hasErrors && (
              <Alert variant="info">
                <p> {t("pages.gateway.documentation.noDocumentation")}</p>
              </Alert>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
};

export default OpenapiDocPage;
