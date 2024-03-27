import { Card } from "~/design/Card";
import Header from "./Header";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { NoPermissions } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { useRoute } from "~/api/gateway/query/getRoutes";

const RoutesDetailPage = () => {
  const { routePath } = pages.gateway.useParams();
  const { data, isAllowed, isFetched, noPermissionMessage } = useRoute({
    routePath: routePath ?? "",
    enabled: !!routePath,
  });

  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  if (!data) return null;

  return (
    <>
      {data.path && <LogStreamingSubscriber route={data.path} />}
      <Header />
    </>
  );
};

export default RoutesDetailPage;
