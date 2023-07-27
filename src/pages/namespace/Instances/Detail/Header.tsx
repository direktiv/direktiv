import { Box } from "lucide-react";
import { FC } from "react";
import { useInstanceDetails } from "~/api/instances/query/details";

const Header: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useInstanceDetails({ instanceId });

  if (!data) return null;

  console.log(data);

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <Box />
    </div>
  );
};

export default Header;
