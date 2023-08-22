import ActivityLog from "./ActivityLog";
import { MirrorActivityLogSubscriber } from "~/api/tree/query/mirrorActivity";
import { pages } from "~/util/router/pages";

const Logs = () => {
  const { activity } = pages.mirror.useParams();

  if (!activity) return null;

  return (
    <>
      <MirrorActivityLogSubscriber activityId={activity} />
      <ActivityLog activityId={activity} />
    </>
  );
};

export default Logs;
