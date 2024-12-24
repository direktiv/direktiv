import { Card } from "~/design/Card";
import { RapiDoc } from "~/design/RapiDoc";
import { useDocs } from "~/api/gateway/query/getDocs";
import { useTranslation } from "react-i18next";

const GatewayDocsPage = () => {
  const { t } = useTranslation();
  const { data, isError } = useDocs();

  if (isError) {
    return <p>Error</p>;
  }
  return (
    <Card className="m-5">
      <div className="flex justify-end gap-5 p-2">
        <RapiDoc spec={data?.data ?? {}} />
      </div>
    </Card>
  );
};

export default GatewayDocsPage;
