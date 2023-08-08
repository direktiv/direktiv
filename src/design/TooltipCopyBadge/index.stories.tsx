import TooltipCopyBadge from ".";
import { TooltipProvider } from "../Tooltip";

export default {
  title: "Components/TooltipCopyBadge",
};

export const Default = () => (
  <div>
    <TooltipProvider>
      <TooltipCopyBadge
        value="some-rather-a-bit-too-long-text"
        displayValue="short-text"
      />
    </TooltipProvider>
  </div>
);
