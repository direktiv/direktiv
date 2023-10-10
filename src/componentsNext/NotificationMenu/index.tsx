import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { FC } from "react";
import Notification from "~/design/Notification";

/*
// Vorher:
const NotificationMenu: FC = () => (
*/

// Nachher:
const NotificationMenu = ({ className }: { className: string }) => (
  <div className="self-end text-right">
    <Popover>
      <PopoverTrigger>
        <Notification />
      </PopoverTrigger>
      <PopoverContent align="end" className="p-4">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
  </div>
);

export default NotificationMenu;
