import { useMirrorInfo } from "~/api/tree/query/mirrorInfo";
import { useTranslation } from "react-i18next";

const Activities = () => {
  const { data } = useMirrorInfo();
  const { t } = useTranslation();

  const activities = data?.activities.results;

  if (!activities) return; // TODO: Render NoResults component?

  return (
    <ul>
      {activities.map((activity) => (
        <li key={activity.id}>{activity.id}</li>
      ))}
    </ul>
  );
};

export default Activities;
