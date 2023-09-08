import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import RefreshButton from "~/design/RefreshButton";
import { useTranslation } from "react-i18next";

const RefetchButton = ({
  disabled,
  onClick,
}: {
  disabled: boolean;
  onClick: () => void;
}) => {
  const { t } = useTranslation();
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <RefreshButton
            icon
            size="sm"
            variant="ghost"
            disabled={disabled}
            onClick={onClick}
          />
        </TooltipTrigger>
        <TooltipContent>
          {t(`pages.monitoring.instances.updateTooltip`)}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

export default RefetchButton;
