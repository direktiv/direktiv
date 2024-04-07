import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import Badge from "~/design/Badge";
import { FC } from "react";
import { RouteSchemaType } from "~/api/gateway/schema";
import { useTranslation } from "react-i18next";

type AllowAnonymousProps = {
  methods: RouteSchemaType["methods"];
};

const methodDisplayCount = 2;

export const Methods: FC<AllowAnonymousProps> = ({ methods }) => {
  const { t } = useTranslation();
  const numberOfMethods = methods?.length ?? 0;
  const numberOfHiddenMethods = numberOfMethods - methodDisplayCount;
  const needsTooltip = numberOfHiddenMethods > 0;

  const methodsToDisplay = methods?.slice(0, methodDisplayCount) ?? [];
  const methodsBehindTooltip = methods?.slice(methodDisplayCount) ?? [];

  return (
    <div className="flex gap-1">
      {numberOfMethods === 0 && (
        <Badge variant="secondary">
          {t("pages.gateway.routes.row.methods.none")}
        </Badge>
      )}
      {methodsToDisplay.map((method) => (
        <Badge key={method} variant="outline">
          {method}
        </Badge>
      ))}
      {needsTooltip && (
        <HoverCard>
          <HoverCardTrigger>
            <Badge variant="outline">
              {t("pages.gateway.routes.row.methods.tooltipLabel", {
                count: numberOfHiddenMethods,
              })}
            </Badge>
          </HoverCardTrigger>
          <HoverCardContent
            align="center"
            side="right"
            className="grid gap-2 p-2"
          >
            {methodsBehindTooltip.map((method) => (
              <Badge key={method} variant="outline">
                {method}
              </Badge>
            ))}
          </HoverCardContent>
        </HoverCard>
      )}
    </div>
  );
};
