import { Card } from "~/design/Card";
import Header from "./Header";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import Logs from "./Logs";
import { NoPermissions } from "~/design/Table";
import { twMergeClsx } from "~/util/helpers";
import { useParams } from "@tanstack/react-router";
import { useRoute } from "~/api/gateway/query/getRoutes";

const RoutesDetailPage = () => {
  const { _splat } = useParams({ strict: false });
  const { data, isAllowed, isFetched, noPermissionMessage } = useRoute({
    routePath: _splat ?? "",
    enabled: !!_splat,
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
    <div className="grid grow grid-rows-[auto_1fr]">
      <LogStreamingSubscriber
        route={data.spec["x-direktiv-config"]?.path}
        enabled={!!data.spec["x-direktiv-config"]?.path}
      />
      <Header />
      <div
        className={twMergeClsx(
          "grid grow gap-5 p-5",
          "grid-rows-[calc(100vh-20rem)]",
          "sm:grid-rows-[calc(100vh-18rem)]",
          "lg:grid-rows-[calc(100vh-13rem)]"
        )}
      >
        <Card className="relative grid grid-rows-[auto,1fr,auto] p-5">
          <Logs path={data.spec["x-direktiv-config"]?.path} />
        </Card>
      </div>
    </div>
  );
};

export default RoutesDetailPage;
