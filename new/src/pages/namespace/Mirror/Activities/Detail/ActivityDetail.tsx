import Header from "./Header";
import Logs from "./Logs";

const ActivityDetail = ({ activityId }: { activityId: string }) => (
  <>
    <Header activityId={activityId} />
    <Logs activityId={activityId} />
  </>
);

export default ActivityDetail;
