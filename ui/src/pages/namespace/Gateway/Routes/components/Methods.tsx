import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import { ConditionalWrapper } from "~/util/helpers";
import { FC } from "react";
import { RouteMethod } from "~/api/gateway/schema";
import { useTranslation } from "react-i18next";

type AllowAnonymousProps = {
  methods: RouteMethod[];
};

export const Methods: FC<AllowAnonymousProps> = ({ methods }) => {
  const { t } = useTranslation();
  const numberOfMethods = methods.length;

  return (
    <ConditionalWrapper
      condition={numberOfMethods > 0}
      wrapper={(children) => (
        <HoverCard>
          <HoverCardTrigger>{children}</HoverCardTrigger>
          <HoverCardContent align="center" side="right" className="p-1">
            <Table>
              <TableBody>
                {methods.map((method) => (
                  <TableRow key={method}>
                    <TableCell>{method}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </HoverCardContent>
        </HoverCard>
      )}
    >
      <Badge variant="outline">
        {t("pages.gateway.routes.row.methods.tooltipLabel", {
          count: numberOfMethods,
        })}
      </Badge>
    </ConditionalWrapper>
  );
};
