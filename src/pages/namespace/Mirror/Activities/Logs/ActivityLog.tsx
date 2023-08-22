import { useMirrorActivityLog } from "~/api/tree/query/mirrorActivity";

const ActivityLog = ({ activityId }: { activityId: string }) => {
  const { data } = useMirrorActivityLog({ activityId });

  const logItems = data?.results;

  if (!logItems) return null;

  return (
    <ul>
      {logItems.map((item, index) => (
        <li key={index}>{item.msg}</li>
      ))}
    </ul>
  );
};

export default ActivityLog;
