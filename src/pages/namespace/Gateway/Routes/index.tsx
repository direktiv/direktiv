import GatewayTable from "./Table";
import { Network } from "lucide-react";
import RefreshButton from "~/design/RefreshButton";
import { useRoutes } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();

  const { isFetching, refetch } = useRoutes();

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
      <GatewayTable />
    </div>
  );
};

export default GatewayPage;
