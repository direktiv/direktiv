import { FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { Link } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type ButtonWithChildren = PropsWithChildren & {
  className?: string;
};

export const NotificationHasresultsButton: FC<ButtonWithChildren> = ({
  children,
  className,
}) => {
  const namespace = useNamespace();
  const { icon: Icon } = pages.settings;
  useTranslation();

  if (!namespace) return null;

  return (
    <Button className="" variant="outline" isAnchor asChild>
      <Link
        className=""
        to={pages.settings.createHref({
          namespace,
        })}
      >
        <Icon aria-hidden="true" />
        {children}
      </Link>
    </Button>
  );
};

export const NotificationNoresults: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationHasresultsTitle: FC<PropsWithChildren> = ({
  children,
}) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationHasresultsText: FC<PropsWithChildren> = ({
  children,
}) => (
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
