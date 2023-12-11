import { Card } from "~/design/Card";
import GatewayTable from "./Table";
import RefreshButton from "~/design/RefreshButton";
import { useRoutes } from "~/api/gateway/query/getRoutes";

const GatewayPage = () => {
  const { isFetching, refetch } = useRoutes();

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
        />
      </div>
      <GatewayTable />
    </Card>
  );
};

export default GatewayPage;
