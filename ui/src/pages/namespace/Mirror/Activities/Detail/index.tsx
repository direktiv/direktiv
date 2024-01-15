import ActivityDetail from "./ActivityDetail";
import { Card } from "~/design/Card";
import { MirrorActivityLogSubscriber } from "~/api/tree/query/mirrorActivity";
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
      <MirrorActivityLogSubscriber activityId={activity} />
      <ActivityDetail activityId={activity} />
    </>
  );
};

export default Logs;
