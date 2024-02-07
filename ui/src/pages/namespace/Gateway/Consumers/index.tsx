import { Card } from "~/design/Card";
import ConsumerTable from "./Table";
import RefreshButton from "~/design/RefreshButton";
import { useConsumers } from "~/api/gateway/query/getConsumers";
import { useTranslation } from "react-i18next";

const ConsumerPage = () => {
  const { isFetching, refetch } = useConsumers();
  const { t } = useTranslation();
  return (
    <Card className="m-5">
      <div className="flex justify-end gap-5 p-2">
        <RefreshButton
          icon
          variant="outline"
          disabled={isFetching}
          onClick={() => {
            refetch();
          }}
          aria-label={t("pages.gateway.consumer.refetchLabel")}
        />
      </div>
      <ConsumerTable />
    </Card>
  );
};

export default ConsumerPage;
