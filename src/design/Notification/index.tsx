import { Bell } from "lucide-react";

const Notification = ({ hasMessage }: { hasMessage?: boolean }): JSX.Element =>
  hasMessage ? (
    <div className="relative h-6 w-6">
      <Bell className="relative" />
      <div className="absolute top-0 right-0 rounded-full border-2 border-white bg-danger-10 p-1 dark:border-black dark:bg-danger-dark-10"></div>
    </div>
  ) : (
    <Bell className="relative" />
  );

export default Notification;
