import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import { ConditionalWrapper } from "~/util/helpers";
import { FC } from "react";
import { RouteSchemaType } from "~/api/gateway/schema";
import { useTranslation } from "react-i18next";

type PluginCountProps = {
  number: number;
  type: keyof RouteSchemaType["plugins"];
};

const PluginCount: FC<PluginCountProps> = ({ type, number }) => {
  const { t } = useTranslation();
  return number > 0 ? (
    <TableRow>
      <TableCell>
        {t("pages.gateway.routes.row.plugin.countType", {
          count: number,
          type,
        })}
      </TableCell>
    </TableRow>
  ) : null;
};

type PluginsProps = {
  plugins: RouteSchemaType["plugins"];
};

const Plugins: FC<PluginsProps> = ({ plugins }) => {
  const numberOfInboundPlugins = plugins.inbound?.length ?? 0;
  const numberOfAuthPlugins = plugins.auth?.length ?? 0;
  const numberOfOutboundPlugins = plugins.outbound?.length ?? 0;
  const numberOfTargetPlugin = plugins.target ? 1 : 0;

  const numberOfPlugins =
    numberOfInboundPlugins +
    numberOfAuthPlugins +
    numberOfOutboundPlugins +
    numberOfTargetPlugin;

  const { t } = useTranslation();

  return (
    <ConditionalWrapper
      condition={numberOfPlugins > 0}
      wrapper={(children) => (
        <HoverCard>
          <HoverCardTrigger>{children}</HoverCardTrigger>
          <HoverCardContent align="center" side="left" className="p-1">
            <Table>
              <TableBody>
                <PluginCount number={numberOfInboundPlugins} type="inbound" />
                <PluginCount number={numberOfOutboundPlugins} type="outbound" />
                <PluginCount number={numberOfAuthPlugins} type="auth" />
                <PluginCount number={numberOfTargetPlugin} type="target" />
              </TableBody>
            </Table>
          </HoverCardContent>
        </HoverCard>
      )}
    >
      <div>
        {t("pages.gateway.routes.row.plugin.countAll", {
          count: numberOfPlugins,
        })}
      </div>
    </ConditionalWrapper>
  );
};

export default Plugins;
