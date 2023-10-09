import { Bell } from "lucide-react";

const Notification = ({
  className,
  hasMessage,
}: {
  className?: string;
  hasMessage?: boolean;
}): JSX.Element =>
  hasMessage ? (
    <div className="relative h-6 w-6">
      <Bell className="relative" />
      <div className="absolute top-0 right-0 rounded-full bg-danger-10 p-1"></div>
    </div>
  ) : (
    <Bell />
  );

export default Notification;
