import Alert from "~/design/Alert";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import { RapiDoc } from "~/design/RapiDoc";
import { useInfo } from "~/api/gateway/query/getInfo";
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
  const { data } = useInfo({
    expand: true,
  });
  const { t } = useTranslation();

  const info = data?.data as DocumentationInfo | undefined;
  const { spec, errors } = info || { spec: null, errors: [] };

  const hasErrors = errors?.length > 0;
  const hasSpec = spec && spec.paths && Object.keys(spec.paths).length > 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      {hasErrors && (
        <Alert variant="error">
          <pre>{JSON.stringify(errors, null, 2)}</pre>
        </Alert>
      )}
      {hasSpec && (
        <>
          <Card className="w-full">
            <div className="flex flex-col">
              <RapiDoc spec={spec} />
            </div>
          </Card>

          <div className="flex justify-end">
            <CopyButton
              value={spec ? JSON.stringify(spec, null, 2) : ""}
              buttonProps={{
                variant: "outline",
                className: "w-36",
              }}
            >
              {(copied) =>
                copied
                  ? t("pages.gateway.documentation.copied")
                  : t("pages.gateway.documentation.copySpec")
              }
            </CopyButton>
          </div>
        </>
      )}
      {!hasSpec && !hasErrors && (
        <Card className="w-full">
          <div className="flex flex-col">
            <Alert variant="info">
              <p>{t("pages.gateway.documentation.noDocumentation")}</p>
            </Alert>
          </div>
        </Card>
      )}
    </div>
  );
};

export default OpenapiDocPage;
