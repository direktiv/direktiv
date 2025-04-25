import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { PropsWithChildren } from "react";

type ErrorProps = PropsWithChildren<{
  value: string;
}>;

export const Error = ({ value, children }: ErrorProps) => (
  <TooltipProvider>
    <Tooltip>
      <TooltipTrigger>
        <span className="bg-danger-4 text-danger-11 dark:bg-danger-dark-4 dark:text-danger-dark-11">
          {value}
        </span>
      </TooltipTrigger>
      <TooltipContent className="w-[450px]">{children}</TooltipContent>
    </Tooltip>
  </TooltipProvider>
);
