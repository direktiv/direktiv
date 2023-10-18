import { FC, PropsWithChildren } from "react";
import { Loader2, Settings } from "lucide-react";

import Button from "~/design/Button";
import { Link } from "react-router-dom";

type ButtonWithChildren = PropsWithChildren & {
  className?: string;
  linkTo?: string;
};

export const NotificationButton: FC<ButtonWithChildren> = ({
  children,
  className,
  linkTo,
}) => {
  // delete probably?
  if (!linkTo) {
    linkTo = "#";
  }
  //

  return (
    <Button variant="outline" isAnchor asChild>
      <Link to={linkTo}>
        <Settings aria-hidden="true" />
        {children}
      </Link>
    </Button>
  );
};

export const NotificationTitle: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationText: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationLoading: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex items-center">
    <Loader2 className="h-5 animate-spin" />
    <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
      {children}
    </div>{" "}
  </div>
);
