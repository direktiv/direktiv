import Badge from "~/design/Badge";
import { Box } from "lucide-react";
import { FC } from "react";
import { statusToBadgeVariant } from "../utils";
import { useInstanceDetails } from "~/api/instances/query/details";

const Header: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useInstanceDetails({ instanceId });

  if (!data) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <Box className="h-5" /> {data.instance.id.slice(0, 8)}
        </h3>
        <Badge variant={statusToBadgeVariant(data.instance.status)}>
          {data.instance.status}
        </Badge>
      </div>
    </div>
  );
};

export default Header;
