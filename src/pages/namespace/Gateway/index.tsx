import { Card } from "~/design/Card";
import { Network } from "lucide-react";
import RefreshButton from "~/design/RefreshButton";
import { useGatewayList } from "~/api/gatewawy/query/get";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();
  const { data, isFetching, refetch } = useGatewayList();
  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex">
        <h3 className="flex grow items-center gap-x-2 font-bold">
          <Network className="h-5" />
          {t("pages.gateway.title")}
        </h3>
        <RefreshButton
          icon
          variant="outline"
          disabled={isFetching}
          onClick={() => {
            refetch();
          }}
        />
      </div>
      <Card className="p-5 text-sm">
        <pre>{data && JSON.stringify(data)}</pre>
      </Card>
    </div>
  );
};

export default GatewayPage;
