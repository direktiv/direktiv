import { ActivitySquare } from "lucide-react";
import { Card } from "~/design/Card";
import { useTranslation } from "react-i18next";

const InstancesPage = () => {
  const { t } = useTranslation();
  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <ActivitySquare className="h-5" />
        {t("pages.monitoring.title")}
      </h3>
      <Card className="p-5"></Card>
    </div>
  );
};

export default InstancesPage;
