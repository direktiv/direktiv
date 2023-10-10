import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Notification from "~/design/Notification";
import { twMergeClsx } from "~/util/helpers";
import { useState } from "react";

interface NotificationMenuProps {
  className?: string;
}

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const [hasMessage, setHasMessages] = useState(false);
  return (
    <div className={twMergeClsx("self-end text-right", className)}>
      <Popover>
        <PopoverTrigger>
          <Notification hasMessage={true} />
        </PopoverTrigger>
        <PopoverContent align="end" className="p-4">
          Place content for the popover here.
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default NotificationMenu;
