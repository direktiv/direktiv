import Header from "./Header";
import Logs from "./Logs";

const ActivityDetail = ({ activityId }: { activityId: string }) => (
  <div className="flex grow flex-col">
    <Header activityId={activityId} />
    <Logs activityId={activityId} />
  </div>
);

export default ActivityDetail;
