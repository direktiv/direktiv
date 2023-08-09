import { InstanceSchemaType } from "~/api/instances/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { TooltipProvider } from "~/design/Tooltip";

export const InstanceRow = ({ instance }: { instance: InstanceSchemaType }) => (
  <TooltipProvider>
    <div key={instance.id}>
      {instance.as}{" "}
      <TooltipCopyBadge value={instance.id} variant="outline">
        {instance.id.slice(0, 8)}
      </TooltipCopyBadge>
    </div>
  </TooltipProvider>
);
