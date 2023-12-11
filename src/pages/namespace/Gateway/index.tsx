import GatewayTable from "./Table";
import { Network } from "lucide-react";
import RefreshButton from "~/design/RefreshButton";
import useIsGatewayAvailable from "~/hooksNext/useIsGatewayAvailable";
import { useRoutes } from "~/api/gateway/query/getEndpoints";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();
  const isGatewayAvailable = useIsGatewayAvailable();
  const { isFetching, refetch } = useRoutes({
    enabled: !!isGatewayAvailable,
  });

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex">
        <h3 className="flex grow items-center gap-x-2 font-bold">
          <Network className="h-5" />
          {t("pages.gateway.title")}
        </h3>
        {isGatewayAvailable && (
          <RefreshButton
            icon
            variant="outline"
            disabled={isFetching}
            onClick={() => {
              refetch();
            }}
          />
        )}
      </div>
      <GatewayTable />
    </div>
  );
};

export default GatewayPage;
