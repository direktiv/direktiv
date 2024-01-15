import TooltipCopyBadge from ".";
import { TooltipProvider } from "../Tooltip";

export default {
  title: "Components/TooltipCopyBadge",
};

export const Default = () => (
  <div className="flex gap-2">
    <TooltipProvider>
      <TooltipCopyBadge value="some-rather-a-bit-too-long-text" icon="complete">
        short-text
      </TooltipCopyBadge>
      <TooltipCopyBadge
        variant="secondary"
        value="some-rather-a-bit-too-long-text"
        icon="pending"
      >
        short-text
      </TooltipCopyBadge>
      <TooltipCopyBadge
        variant="outline"
        value="some-rather-a-bit-too-long-text"
      >
        short-text
      </TooltipCopyBadge>
      <TooltipCopyBadge
        value="some-rather-a-bit-too-long-text"
        variant="destructive"
        icon="failed"
      >
        short-text
      </TooltipCopyBadge>
      <TooltipCopyBadge
        value="some-rather-a-bit-too-long-text"
        variant="success"
        icon="complete"
      >
        short-text
      </TooltipCopyBadge>
    </TooltipProvider>
  </div>
);
