import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import Alert from "~/design/Alert";
import Badge from "~/design/Badge";
import { FC } from "react";
import { useTranslation } from "react-i18next";

export type ErrorBadgeProps = {
  error: string;
};
const ErrorBadge: FC<ErrorBadgeProps> = ({ error }) => {
  const { t } = useTranslation();

  return error ? (
    <HoverCard>
      <HoverCardTrigger className="inline-flex">
        <Badge variant="destructive">{t("pages.gateway.row.error")}</Badge>
      </HoverCardTrigger>
      <HoverCardContent asChild noBackground>
        <Alert variant="error" className="w-96 whitespace-pre-wrap break-all">
          {error}
        </Alert>
      </HoverCardContent>
    </HoverCard>
  ) : null;
};

export default ErrorBadge;
