import { Tooltip, TooltipContent, TooltipTrigger } from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { ComponentProps } from "react";
import CopyButton from "~/design/CopyButton";

type TooltipCopyBadge = ComponentProps<typeof Badge> & {
  value: string;
};

const TooltipCopyBadge = ({ value, ...props }: TooltipCopyBadge) => (
  <Tooltip>
    <TooltipTrigger data-testid="tooltip-copy-trigger">
      <Badge {...props} />
    </TooltipTrigger>
    <TooltipContent
      data-testid="tooltip-copy-content"
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
