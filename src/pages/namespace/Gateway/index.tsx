import GatewayTable from "./Table";
import { Network } from "lucide-react";
import NoGateway from "./NoGateway";
import RefreshButton from "~/design/RefreshButton";
import { useGatewayList } from "~/api/gateway/query/get";
import useIsGatewayAvailable from "~/hooksNext/useIsGatewayAvailable";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();
  const isGatewayAvailable = useIsGatewayAvailable();
  const { isFetching, refetch } = useGatewayList({
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
      {isGatewayAvailable ? <GatewayTable /> : <NoGateway />}
    </div>
  );
};

export default GatewayPage;
