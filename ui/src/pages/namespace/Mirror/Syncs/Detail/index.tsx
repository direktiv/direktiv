import ActivityDetail from "./ActivityDetail";
import { Card } from "~/design/Card";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { NoPermissions } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { useMirrorActivity } from "~/api/tree/query/mirrorInfo";

const Logs = () => {
  const { activity } = pages.mirror.useParams();
  const { isAllowed, noPermissionMessage, isFetched } = useMirrorActivity({
    id: activity ?? "",
  });

  if (!activity) return null;
  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <>
      <LogStreamingSubscriber activity={activity} />
      <ActivityDetail syncId={activity} />
    </>
  );
};

export default Logs;
