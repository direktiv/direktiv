import { FC, PropsWithChildren } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { Info } from "lucide-react";

const InfoTooltip: FC<PropsWithChildren> = ({ children }) => (
  <TooltipProvider>
    <Tooltip delayDuration={100}>
      <TooltipTrigger type="button">
        <Info size={16} className="ml-2 text-gray-11" />
      </TooltipTrigger>
      <TooltipContent className="max-w-md text-left">{children}</TooltipContent>
    </Tooltip>
  </TooltipProvider>
);

export default InfoTooltip;
