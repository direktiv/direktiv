import { Card } from "../../../../../../design/Card";
import { FC } from "react";
import { Network } from "lucide-react";

const TrafficShaping: FC = () => (
  <>
    <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
      <Network />
      Traffic Shaping
    </h3>
    <Card className="flex gap-x-3 p-4"></Card>
  </>
);

export default TrafficShaping;
