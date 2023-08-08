import { Tooltip, TooltipContent, TooltipTrigger } from "~/design/Tooltip";

import Badge from "~/design/Badge";
import CopyButton from "~/design/CopyButton";

const TooltipCopyBadge = ({
  value,
  displayValue,
}: {
  value: string;
  displayValue: string;
}) => (
  <Tooltip>
    <TooltipTrigger data-testid={`instance-row-id-${value}`}>
      <Badge variant="outline">{displayValue}</Badge>
    </TooltipTrigger>
    <TooltipContent
      data-testid={`instance-row-id-full-${value}`}
      className="flex gap-2 align-middle"
    >
      {value}
      <CopyButton
        value={value}
        buttonProps={{
          size: "sm",
          onClick: (e) => {
            e.stopPropagation();
          },
        }}
      />
    </TooltipContent>
  </Tooltip>
);

export default TooltipCopyBadge;
