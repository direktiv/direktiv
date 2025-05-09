import { ButtonHTMLAttributes, PropsWithChildren, forwardRef } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

type ErrorProps = ButtonHTMLAttributes<HTMLButtonElement> &
  PropsWithChildren<{
    value: string;
  }>;

export const Error = forwardRef<HTMLButtonElement, ErrorProps>(
  ({ value, children, ...props }, ref) => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger ref={ref} asChild {...props}>
          <span className="bg-danger-4 text-danger-11 dark:bg-danger-dark-4 dark:text-danger-dark-11">
            {value}
          </span>
        </TooltipTrigger>
        <TooltipContent className="w-[450px]">{children}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
);

Error.displayName = "Error";
