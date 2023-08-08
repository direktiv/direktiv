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
    <TooltipTrigger data-testid={`tooltip-copy-badge-${value}`}>
      <Badge variant="outline">{displayValue}</Badge>
    </TooltipTrigger>
    <TooltipContent
      data-testid={`tooltip-copy-badge-content-${value}`}
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
